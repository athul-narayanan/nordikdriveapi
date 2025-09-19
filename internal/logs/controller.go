package logs

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type LogController struct {
	LogService *LogService
}

func (lc *LogController) GetLogs(c *gin.Context) {
	var input LogFilterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logs, total, totalPages, err := lc.LogService.GetLogs(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        logs,
		"page":        input.Page,
		"page_size":   input.PageSize,
		"total":       total,
		"total_pages": totalPages,
	})
}
