package main

import (
	"note-system/config"
	"note-system/internal/model"
	"os"

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

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"msg": "Hello Gin!"})
	})
	router.Run(":" + cfg.Server.Port)
}
