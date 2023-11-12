//go:build !windows

package main

import (
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
)

func initServer(address string, router *gin.Engine) server {
	s := endless.NewServer(address, router)
	s.ReadHeaderTimeout = 1800 * time.Second
	s.WriteTimeout = 1800 * time.Second
	s.MaxHeaderBytes = 1 << 20
	return s
}
