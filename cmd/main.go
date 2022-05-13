package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	//router.GET("/", funcHandler)
	router.Run(":8000")

}
