package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

const (
	Port = ":9333"
	// ProxyUrl 如果不使用代理，设为空""即可
	ProxyUrl = "http://127.0.0.1:7890"
	// DisableHistory 默认true不开启网页历史记录
	DisableHistory = true
)

func main() {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.OPTIONS("/v1/chat/completions", optionsHandler)
	router.POST("/v1/chat/completions", chatCompletions)
	router.POST("/v1/chat/dalle", dalle)

	s := initServer(Port, router)
	fmt.Println(s.ListenAndServe().Error())
}
