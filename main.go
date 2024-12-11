package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 定义用户的结构体
type User struct {
	Id       int    `json:"id" gorm:"primaryKey"`
	Username string `json:"username" gorm:"unique"`
	Password string `json:"password"`
}

// 定义问题的结构体
type Question struct {
	Id      int      `json:"id" gorm:"primaryKey"`
	Title   string   `json:"title"`
	Content string   `json:"content"`
	UserId  int      `json:"user_id"`
	Answers []Answer `json:"answers" gorm:"foreignKey:QuestionId"`
}

// 定义回答的结构体
type Answer struct {
	Id         int    `json:"id" gorm:"primaryKey"`
	Content    string `json:"content"`
	UserId     int    `json:"user_id"`
	QuestionId int    `json:"question_id"`
	IsBest     bool   `json:"is_best"`
}

var db *gorm.DB

func main() {
	// 创建 Gin 引擎实例，只需要一个
	r := gin.Default()

	// 连接数据库，此处使用的是Sqlite，注意不同的数据库格式不一样
	var err error
	db, err = gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}

	// 数据库迁移
	db.AutoMigrate(&User{}, &Question{}, &Answer{})

	// 用户注册和登录
	r.POST("/register", Register)
	r.POST("/login", Login)

	// 问题和回答
	v := r.Group("/v1")
	{
		v.POST("/questions", CreateQuestion)                // 提出问题
		v.GET("/questions", GetQuestions)                   // 获取所有问题
		v.POST("/questions/:id/answers", CreateAnswer)      // 回答问题
		v.POST("/questions/:id/best_answer", SetBestAnswer) // 设置最优回答
	}

	// 启动服务器，需要放在main函数里
	r.Run(":8080") // 监听在 8080 端口
}

// 定义注册的函数
func Register(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// 保存用户到数据库
	db.Create(&user)
	c.JSON(201, user)
}

// 定义登录的函数
func Login(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// 验证用户
	var dbUser User
	db.Where("username = ? AND password = ?", user.Username, user.Password).First(&dbUser)
	if dbUser.Id == 0 {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}
	c.JSON(200, dbUser)
}

// 创建问题的函数
func CreateQuestion(c *gin.Context) {
	var question Question
	if err := c.ShouldBindJSON(&question); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	db.Create(&question)
	c.JSON(201, question)
}

// 获取所有问题
func GetQuestions(c *gin.Context) {
	var questions []Question
	db.Preload("Answers").Find(&questions) // 预加载问题的答案
	c.JSON(200, questions)
}

// 处理回答问题
func CreateAnswer(c *gin.Context) {
	var answer Answer
	if err := c.ShouldBindJSON(&answer); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	db.Create(&answer)
	c.JSON(201, answer)
}

// 设置最优回答
func SetBestAnswer(c *gin.Context) {
	questionId := c.Param("id")
	var answer Answer
	if err := c.ShouldBindJSON(&answer); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	// 设置最优回答，其他回答标记为非最优
	db.Model(&Answer{}).Where("id = ?", answer.Id).Update("is_best", true)
	db.Model(&Answer{}).Where("question_id = ? AND id != ?", questionId, answer.Id).Update("is_best", false)
	c.JSON(200, gin.H{"message": "Best answer set!"})
}
