package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	
	router.GET("/objects/:id", func(c *gin.Context) {

		id := c.Param("id")
		c.Status(200)

		// retrieve the data if it exists

		c.JSON(http.StatusOK, gin.H{"id": id})
	})

	router.Run()
}
