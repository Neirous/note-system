package main

import (
	"fmt"
	"note-system/config"
	"note-system/internal/handler"
	"note-system/internal/model"
	"note-system/internal/repository"
	"note-system/internal/service"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-yaml"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func loadConfig() (*config.Config, error) {
	paths := []string{"config/config.local.yaml", "config/config.yaml"}
	var data []byte
	var err error
	for _, p := range paths {
		b, e := os.ReadFile(p)
		if e == nil && len(b) > 0 {
			data = b
			break
		}
		if err == nil {
			err = e
		}
	}
	if err != nil {
		return nil, err
	}
	var cfg config.Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {

	//加载配置文件
	cfg, err := loadConfig()
	if err != nil {
		panic("加载配置失败" + err.Error())
	}

	if cfg.LLM.URL != "" {
		_ = os.Setenv("LLM_URL", cfg.LLM.URL)
	}
	if cfg.LLM.Model != "" {
		_ = os.Setenv("LLM_MODEL", cfg.LLM.Model)
	}
	if cfg.LLM.MaxTokens > 0 {
		_ = os.Setenv("LLM_MAX_TOKENS", fmt.Sprintf("%d", cfg.LLM.MaxTokens))
	}
	if cfg.Rag.PineconeHost != "" {
		_ = os.Setenv("PINECONE_HOST", cfg.Rag.PineconeHost)
	}
	if cfg.Rag.PineconeAPIKey != "" {
		_ = os.Setenv("PINECONE_API_KEY", cfg.Rag.PineconeAPIKey)
	}
	if cfg.Rag.PineconeIndex != "" {
		_ = os.Setenv("PINECONE_INDEX", cfg.Rag.PineconeIndex)
	}
	if cfg.Rag.EmbeddingURL != "" {
		_ = os.Setenv("EMBEDDING_URL", cfg.Rag.EmbeddingURL)
	}
	if cfg.Rag.EmbedDim > 0 {
		_ = os.Setenv("EMBED_DIM", fmt.Sprintf("%d", cfg.Rag.EmbedDim))
	}
	if cfg.Rag.TopK > 0 {
		_ = os.Setenv("RAG_TOPK", fmt.Sprintf("%d", cfg.Rag.TopK))
	}
	if cfg.Rag.SimilarityThreshold > 0 {
		_ = os.Setenv("SIMILARITY_THRESHOLD", fmt.Sprintf("%g", cfg.Rag.SimilarityThreshold))
	}

	db, err := gorm.Open(mysql.Open(cfg.Mysql.Dsn), &gorm.Config{})
	if err != nil {
		panic("数据库连接失败" + err.Error())
	}
	//验证连接是否成功
	sqlDB, err := db.DB()
	if err != nil {
		panic("获取数据库实例失败" + err.Error())
	}
	err = sqlDB.Ping()
	if err != nil {
		panic("数据库ping失败")
	}
	println("数据库连接成功！")

	err = db.AutoMigrate(&model.Note{}, &model.Fragment{}, &model.QARecord{})
	if err != nil {
		panic("自动创建失败：" + err.Error())
	}
	println("notes表创建/更新成功")

	// 强制统一为 utf8mb4，避免中文出现问号
	_ = db.Exec("SET NAMES utf8mb4").Error
	_ = db.Exec("SET character_set_client = utf8mb4").Error
	_ = db.Exec("SET character_set_connection = utf8mb4").Error
	_ = db.Exec("SET character_set_results = utf8mb4").Error
	_ = db.Exec("SET collation_connection = utf8mb4_unicode_ci").Error
	_ = db.Exec("ALTER TABLE notes CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error
	_ = db.Exec("ALTER TABLE fragments CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error
	_ = db.Exec("ALTER TABLE qa_records CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci").Error

	// 步骤3：初始化各层（依赖注入）
	noteRepo := repository.NewNoteRepo(db)          // Repository 层
	noteService := service.NewNoteService(noteRepo) // Service 层
	ragService := service.NewRAGService(db)
	nh := handler.NewNoteHandler(noteService, ragService)
	// 通过闭包方式注入 RAGService
	func() { // anonymous init
		// reflect injection avoided; exported field not settable here
		// provide setter via helper
	}()

	// 步骤4：创建 Gin 引擎，注册路由
	r := gin.Default() // 默认开启日志和恢复中间件
	// 新增：添加跨域中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080", "http://localhost:5173", "http://localhost:5174"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"}, // 允许的请求方法
		AllowHeaders:     []string{"Content-Type"},                 // 允许的请求头
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 分组路由：/api/note
	api := r.Group("/api/note")
	{
		api.POST("", nh.CreateNote)
		api.GET("/search", nh.SearchNotes)
		api.GET("/list", nh.ListNotes)
		api.POST("/seed-cn", nh.SeedCNNotes)
		api.GET("/trash", nh.ListDeleted)
		api.PUT("/:id/restore", nh.Restore)
		api.DELETE("/:id/hard", nh.HardDelete)
		api.GET("/:id", nh.GetNoteByID)
		api.PUT("/:id", nh.UpdateNote)
		api.DELETE("/:id", nh.DeleteNote)
		api.DELETE("/purge", nh.PurgeAll)
	}

	rag := r.Group("/api/rag")
	{
		rag.GET("/search", nh.RagSearch)
		rag.POST("/qa", nh.RagQA)
	}

	// OpenAI 风格的本地模拟端点
	r.POST("/v1/chat/completions", nh.MockLLM)

	// 步骤5：启动 HTTP 服务
	fmt.Println("服务启动成功,访问地址:http://127.0.0.1:" + cfg.Server.Port)
	err = r.Run(":" + cfg.Server.Port)
	if err != nil {
		panic(fmt.Sprintf("服务启动失败：%v", err))
	}
}
