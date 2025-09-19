package community

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, communityService *CommunityService) {
	communityController := &CommunityController{CommunityService: communityService}

	userGroup := r.Group("/communities")
	{
		userGroup.GET("", communityController.GetAllCommunities)
	}

}
