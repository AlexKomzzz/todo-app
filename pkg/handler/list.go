package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Создание handler функций для работы List
func createList(c *gin.Context) {

}

func getAllLists(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Status": "Ok",
	})
}

func getListById(c *gin.Context) {

}

func updateList(c *gin.Context) {

}

func deleteList(c *gin.Context) {

}
