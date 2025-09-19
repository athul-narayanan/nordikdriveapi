package logs

import (
	"nordik-drive-api/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, logService *LogService) {
	logController := &LogController{LogService: logService}

	userGroup := r.Group("/api/logs")
	userGroup.Use(middlewares.AuthMiddleware())
	{
		userGroup.POST("", logController.GetLogs)
	}

}
