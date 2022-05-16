package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type error struct {
	Message string `json:"message"`
}

func newErrorResponse(c *gin.Context, statusCode int, message string) { // Ф-я обработчик ошибки
	logrus.Error(message)
	c.AbortWithStatusJSON(statusCode, error{Message: message})
}
