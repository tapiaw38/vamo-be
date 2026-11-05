package trips

import "github.com/gin-gonic/gin"

func bind(c *gin.Context, input any) bool {
	if err := c.ShouldBindJSON(input); err != nil {
		c.JSON(400, gin.H{"code": "common:invalid-body", "message": "invalid request body"})
		return false
	}
	return true
}
