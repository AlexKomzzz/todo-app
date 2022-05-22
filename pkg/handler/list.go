package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"todo-app"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

var duration time.Duration = 3600 * time.Second

// @Summary Create todo List
// @Security ApiKeyAuth
// @Tags lists
// @Description create todo List
// @ID create-list
// @Accept json
// @Produce json
// @Param input body todo.TodoList true "List info"
// @Success 200 {integer} integer 1
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/lists [post]
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

// @Summary Get All Lists
// @Security ApiKeyAuth
// @Tags lists
// @Description get all lists
// @ID get-all-lists
// @Accept  json
// @Produce  json
// @Success 200 {object} getAllListsResponce
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/lists [get]
func (h *Handler) getAllLists(c *gin.Context) {
	lists := make([]todo.TodoList, 0)
	val, err := h.redisClient.Get(h.ctx, "lists").Result() // Проверяем существует ли ключ "lists" в redis
	if err == redis.Nil {                                  // Если ключа не существует, вытаскиваем данные из postgres и кэшируем в redis

		logrus.Print("Request to Postgres")

		userId, err := getUserId(c) // Определяем ID юзера по токену
		if err != nil {
			return
		}

		lists, err = h.services.TodoList.GetAll(userId) // вытаскиваем списки из БД для определенного пользователя
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		data, err := json.Marshal(lists) // декодируем JSON в слайз байт для дальнейшей записи в redis
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		h.redisClient.Set(h.ctx, "lists", string(data), duration) // Создание записи в redis с ключом "lists"

	} else if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	} else { // Если в redis есть ключ...
		logrus.Print("Request to Redis")
		json.Unmarshal([]byte(val), &lists) // забираем от туда данные и отправляем
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

	var input todo.UpdateListInput
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

// @Summary Delete todo List
// @Security ApiKeyAuth
// @Tags lists
// @Descriprion gelete list by id
// @ID delete-list
// @Accept json
// @Produce json
// @Param id path int true "List Id"
// @Success 200 {integer} integer 1
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/lists/{id} [delete]
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
