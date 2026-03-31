package routes

import "github.com/gin-gonic/gin"

func StartupRouters(r *gin.Engine) {
	CreateUserRouters(r)
}
