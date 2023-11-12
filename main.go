package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"os"
	"strings"
)

const (
	Port = ":9333"
	// ProxyUrl 如果不使用代理，设为空""即可
	ProxyUrl = "http://127.0.0.1:7890"
	// DisableHistory 默认true不开启网页历史记录
	DisableHistory = true
)

var Timer = NewTimerTask()
var DB *gorm.DB
var m = Mysql{
	Path:         "127.0.0.1", // 主机地址127.0.0.1
	Port:         "3306",
	Config:       "charset=utf8mb4&parseTime=True&loc=Local",
	Dbname:       "chatgpt", // 数据库名称
	Username:     "root",    // 用户名
	Password:     "root",    // 密码
	Prefix:       "",
	Singular:     false,
	Engine:       "",
	MaxIdleConns: 10,
	MaxOpenConns: 100,
	LogMode:      "error",
	LogZap:       false,
}

func main() {
	DB = GormMysql()
	//TimerFunc()
	if DB != nil {
		//RegisterTables()
		//程序结束前关闭数据库链接
		db, _ := DB.DB()
		defer func(db *sql.DB) {
			err := db.Close()
			if err != nil {
				fmt.Println("关闭数据库连接失败:", err)
			}
		}(db)
	}
	//select {}
	StartHttpServer()
}

func StartHttpServer() {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.OPTIONS("/v1/chat/completions", optionsHandler)
	router.POST("/v1/chat/completions", chatCompletions)

	s := initServer(Port, router)
	fmt.Println(s.ListenAndServe().Error())
}

func readAccounts() {
	var accounts []MailAccount
	// Read accounts.txt and create a list of accounts
	if _, err := os.Stat("accounts.txt"); err == nil {
		var num uint = 3
		// Each line is a proxy, put in proxies array
		file, _ := os.Open("accounts.txt")
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			// Split by :
			line := strings.Split(scanner.Text(), ":")
			// Create an account
			account := MailAccount{
				Username:       line[0],
				OpenaiPassword: line[1],
				OpenaiStatus:   &num,
			}
			// Append to accounts
			accounts = append(accounts, account)
		}
	}
	for _, account := range accounts {
		err := DB.Create(&account).Error
		if err != nil {
			fmt.Println(account.Username)
			fmt.Println(err)
		}
	}
}
