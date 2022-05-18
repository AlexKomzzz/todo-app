package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"todo-app"

	"github.com/gin-gonic/gin"
)

// Создание handler функций для работы List
func (h *Handler) createList(c *gin.Context) {
	userId, err := getUserId(c) // Определяем ID юзера по токену
	if err != nil {
		return
	}

	var input todo.TodoList
	if err := c.BindJSON(&input); err != nil { // парсим тело запроса в структуру List
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.services.TodoList.Create(userId, input) // Создаем список в базе данных
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{ // Отвечаем ОК, id list
		"id": id,
	})
}

type getAllListsResponce struct { // Структура для использования в ответе
	Data []todo.TodoList `json:"data"`
}

func (h *Handler) getAllLists(c *gin.Context) {
	userId, err := getUserId(c) // Определяем ID юзера по токену
	if err != nil {
		return
	}

	lists, err := h.services.TodoList.GetAll(userId) // вытаскиваем списки из БД для определенного пользователя
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, getAllListsResponce{
		Data: lists,
	})
}

func (h *Handler) getListById(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, "ivalid user id")
		return
	}

	id, err := strconv.Atoi(c.Param("id")) // парсим URL, определяем id списка
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid type list id")
		return
	}

	list, err := h.services.TodoList.GetById(userId, id) // вытаскиваем из БД список по id списка и пользователя
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, list)
}

func (h *Handler) updateList(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, "ivalid user id")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid type list id")
		return
	}

	var input todo.TodoList
	if err := c.BindJSON(&input); err != nil { // парсим тело запроса в структуру List
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	list, err := h.services.TodoList.UpdateById(userId, id, input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, list)
}

func (h *Handler) deleteList(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, "ivalid user id")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid type list id")
		return
	}

	err = h.services.TodoList.DeleteById(userId, id) // Удаляем из таблицы Списков и связывающей таблицы список по id
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Ok": fmt.Sprintf("deleted list by id: %d", id),
	})
}
