package role

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RoleController struct {
	RoleService *RoleService
}

func (rc *RoleController) GetAllRoles(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user ID not found"})
		return
	}

	userID, ok := userIDVal.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
		return
	}

	userRoles, err := rc.RoleService.GetRoleByUser(int(userID))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	uniqueRolesMap := make(map[string]struct{})
	for _, r := range userRoles {
		uniqueRolesMap[r.Role] = struct{}{}
	}

	uniqueRoles := make([]string, 0, len(uniqueRolesMap))
	for role := range uniqueRolesMap {
		uniqueRoles = append(uniqueRoles, role)
	}

	fmt.Println("unique roles", uniqueRoles)

	roles, err := rc.RoleService.GetAllRoles(uniqueRoles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Roles fetched successfully",
		"roles":   roles,
	})
}

func (rc *RoleController) GetRolesByUserId(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user ID not found"})
		return
	}

	userID, ok := userIDVal.(float64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
		return
	}

	userRoles, err := rc.RoleService.GetRolesByUserId(int(userID))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Roles fetched successfully",
		"roles":   userRoles,
	})
}
