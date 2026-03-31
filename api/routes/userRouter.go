package routes

import (
	"big2backend/api/modules"

	"github.com/gin-gonic/gin"
)

func CreateUserRouters(r *gin.Engine) {
	userModule := modules.NewUserModule()
	v1 := r.Group("/api/v1")
	{
		usersGroup := v1.Group("/user")
		{
			// GET /api/v1/users?min_age=20
			usersGroup.GET("/list", userModule.GetUsers)

			// POST /api/v1/users
			usersGroup.POST("/add", userModule.CreateUser)

			// GET /api/v1/users/:id
			usersGroup.GET("/get/:id", userModule.GetUser)

			// 可以在這裡添加更多路由，如 PUT, DELETE
		}
	}
}
