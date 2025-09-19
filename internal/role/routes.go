package role

import (
	"nordik-drive-api/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, roleService *RoleService) {
	roleController := &RoleController{RoleService: roleService}

	userGroup := r.Group("/api/role")
	userGroup.Use(middlewares.AuthMiddleware())
	{
		userGroup.GET("", roleController.GetAllRoles)
		userGroup.GET("/user", roleController.GetRolesByUserId)
	}

}
