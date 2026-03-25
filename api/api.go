package api

import (
	"big2backend/api/middlewares"
	"big2backend/api/routes"
	"fmt"

	"github.com/gin-gonic/gin"
)

func StartAPI() {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middlewares.LoggerMiddleware())

	routes.StartupRouters(r)

	fmt.Println("API server starting on :5000")
	if err := r.Run(":5000"); err != nil {
		panic(err)
	}
}
