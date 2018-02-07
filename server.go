package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"os"
)

func main() {
	router := gin.Default()
	router.LoadHTMLGlob(filepath.Join(os.Getenv("GOPATH"), "src/website/views/**/*"))
	router.Static("/css", "src/website/assets/css")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Hello World",
		})
	})
	router.Run(":8080")
}