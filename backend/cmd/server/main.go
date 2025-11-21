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
	//读取config.yaml内容
	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		return nil, err
	}
	//解析yaml内容到Config结构体
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

	err = db.AutoMigrate(&model.Note{})
	if err != nil {
		panic("自动创建失败：" + err.Error())
	}
	println("notes表创建/更新成功")

	// 步骤3：初始化各层（依赖注入）
	noteRepo := repository.NewNoteRepo(db)             // Repository 层
	noteService := service.NewNoteService(noteRepo)    // Service 层
	noteHandler := handler.NewNoteHandler(noteService) // Handler 层

	// 步骤4：创建 Gin 引擎，注册路由
	r := gin.Default() // 默认开启日志和恢复中间件
	// 新增：添加跨域中间件
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080", "http://localhost:5173"}, // 允许前端域名
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},                   // 允许的请求方法
		AllowHeaders:     []string{"Content-Type"},                                   // 允许的请求头
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 分组路由：/api/note
	api := r.Group("/api/note")
	{
		api.POST("", noteHandler.CreateNote)       // 创建笔记
		api.GET("/:id", noteHandler.GetNoteByID)   // 查询单条笔记
		api.PUT("/:id", noteHandler.UpdateNote)    // 更新笔记
		api.DELETE("/:id", noteHandler.DeleteNote) // 删除笔记
		api.GET("/list", noteHandler.ListNotes)    // 分页查询列表
	}

	// 步骤5：启动 HTTP 服务
	fmt.Println("服务启动成功,访问地址:http://127.0.0.1:8080")
	err = r.Run(":8080") // 监听 8080 端口
	if err != nil {
		panic(fmt.Sprintf("服务启动失败：%v", err))
	}
}
