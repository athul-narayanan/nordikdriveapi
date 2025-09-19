package community

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type CommunityController struct {
	CommunityService *CommunityService
}

func (cc *CommunityController) GetAllCommunities(c *gin.Context) {
	communities, err := cc.CommunityService.GetAllCommunities()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Communities fetched successfully",
		"communities": communities,
	})
}
