package handler

import (
	"fmt"
	"net/http"
	"todo-app"

	"github.com/gin-gonic/gin"
)

func (h *Handler) signUp(c *gin.Context) { // Обработчик для регистрации
	var input todo.User

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("invalid input body: %s", err.Error()))
		return
	}

	id, err := h.services.Authorization.CreateUser(input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id,
	})

	// Test
	/*c.JSON(http.StatusOK, gin.H{
		"name":     input.Name,
		"username": input.Username,
		"password": input.Password,
	})*/

}

func (h *Handler) signIn(c *gin.Context) { // Обработчик для аутентификации

}
