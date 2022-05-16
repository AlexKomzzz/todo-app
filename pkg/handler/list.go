package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

/*type Test struct {
	Id string `json:"id"`
}*/

// Создание handler функций для работы List
func (h *Handler) createList(c *gin.Context) {
	/*var id Test
	if err := c.BindJSON(&id); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Id": id.Id,
	})*/
}

func (h *Handler) getAllLists(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"Status": "Ok",
	})
}

func (h *Handler) getListById(c *gin.Context) {

}

func (h *Handler) updateList(c *gin.Context) {

}

func (h *Handler) deleteList(c *gin.Context) {

}
