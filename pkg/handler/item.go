package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"todo-app"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// @Summary Create todo Item
// @Security ApiKeyAuth
// @Tags items
// @Description create todo Item
// @ID create-item
// @Accept json
// @Produce json
// @Param id path int true "List Id"
// @Param input body todo.TodoItem true "Item info"
// @Success 200 {integer} integer 1
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/lists/{id}/items [post]
func (h *Handler) createItem(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	listId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid list id param")
		return
	}

	var input todo.TodoItem
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	id, err := h.services.TodoItem.Create(userId, listId, input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Удаляем список items:listId
	err = h.services.TodoItemCach.HDelete(userId, listId)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id,
	})
}

// @Summary Get All Items
// @Security ApiKeyAuth
// @Tags items
// @Description get all Items
// @ID get-all-items
// @Accept  json
// @Produce  json
// @Param id path int true "List Id"
// @Success 200 {object} []todo.TodoItem
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/lists/{id}/items [get]
func (h *Handler) getAllItems(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	listId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid list id param")
		return
	}

	var items []todo.TodoItem

	// Ищем в кэше ключ items:userId:listId, если его нет, то отправляемся к БД, если есть, то достаем и отправляем
	val, err := h.services.TodoItemCach.HGet(userId, listId, -1)
	if err == redis.Nil { // Если в кэше нет  items, берем из БД

		logrus.Print("Request to Postgres")

		items, err = h.services.TodoItem.GetAll(userId, listId)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		data, err := json.Marshal(items) // Конвертируем структуру в слайз байт
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		// Добавим items в кэш Redis. Используем команду конвейер (Pipeline) для одновременного выполнения команд записи в кэш и установление тайм-аута ключа
		err = h.services.TodoItemCach.HSet(userId, listId, -1, string(data))
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
	} else if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	} else { // если ключ есть в кэше, то отправляем его значение
		logrus.Print("Request to Redis")
		json.Unmarshal([]byte(val), &items)
	}

	c.JSON(http.StatusOK, items)
}

// @Summary Get Item By Id
// @Security ApiKeyAuth
// @Tags items
// @Description get Item by id
// @ID get-item-by-id
// @Accept  json
// @Produce  json
// @Param id path int true "Item Id"
// @Success 200 {object} todo.TodoItem
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/items/{id} [get]
func (h *Handler) getItemById(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	itemId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid list id param")
		return
	}

	var item todo.TodoItem

	// Проверяем нахождение hlists:'usersId'U поля item:'id'
	val, err := h.services.TodoItemCach.HGet(userId, -1, itemId)
	if err == redis.Nil {

		logrus.Print("Request to Postgres")

		item, err = h.services.TodoItem.GetById(userId, itemId)
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		data, err := json.Marshal(item) // Конвертируем структуру в слайз байт
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		// Добавим item в кэш Redis.
		err = h.services.TodoItemCach.HSet(userId, -1, itemId, string(data))
		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

	} else if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	} else {
		logrus.Print("Request to Redis")
		json.Unmarshal([]byte(val), &item)
	}
	c.JSON(http.StatusOK, item)
}

// @Summary Update Item
// @Security ApiKeyAuth
// @Tags items
// @Description update Item
// @ID update-item
// @Accept  json
// @Produce  json
// @Param id path int true "Item Id"
// @Param input body todo.UpdateItemInput true "New item options"
// @Success 200 {object} string
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/items/{id} [put]
func (h *Handler) updateItem(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	var input todo.UpdateItemInput
	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.services.TodoItem.Update(userId, id, input); err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Удаляем все данные из кэша Redis, т.к. у нас нет listId для удаления item:id
	err = h.services.TodoItemCach.Delete(userId)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{"ok"})
}

// @Summary Delete todo Item
// @Security ApiKeyAuth
// @Tags items
// @Descriprion gelete item by id
// @ID delete-item
// @Accept json
// @Produce json
// @Param id path int true "Item Id"
// @Success 200 {integer} string
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /api/items/{id} [delete]
func (h *Handler) deleteItem(c *gin.Context) {
	userId, err := getUserId(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	itemId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid id param")
		return
	}

	err = h.services.TodoItem.Delete(userId, itemId)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Удаляем все данные из кэша Redis, т.к. у нас нет listId для удаления item:id
	err = h.services.TodoItemCach.Delete(userId)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{"ok"})
}
