package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"todo-app"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

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
	err = h.redisClient.HDel(h.ctx, fmt.Sprintf("user:%d", userId), fmt.Sprintf("items:list%d", listId)).Err()
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"id": id,
	})
}

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
	val, err := h.redisClient.HGet(h.ctx, fmt.Sprintf("user:%d", userId), fmt.Sprintf("items:list%d", listId)).Result()
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
		pipe := h.redisClient.Pipeline() // создание конвейра

		pipe.HSetNX(h.ctx, fmt.Sprintf("user:%d", userId), fmt.Sprintf("items:list%d", listId), string(data)) // Кешируем lists в Redis

		pipe.Expire(h.ctx, fmt.Sprintf("user:%d", userId), duration) // Устанавливаем тайм-айт для ключа

		_, err = pipe.Exec(h.ctx) // Выполняем команды конвейера
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
	val, err := h.redisClient.HGet(h.ctx, fmt.Sprintf("user:%d", userId), fmt.Sprintf("item:%d", itemId)).Result()
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

		// Добавим item в кэш Redis. Используем команду конвейер (Pipeline) для одновременного выполнения команд записи в кэш и установление тайм-аута ключа
		pipe := h.redisClient.Pipeline() // создание конвейра

		pipe.HSetNX(h.ctx, fmt.Sprintf("user:%d", userId), fmt.Sprintf("item:%d", itemId), string(data)) // Кешируем lists в Redis

		pipe.Expire(h.ctx, fmt.Sprintf("user:%d", userId), duration) // Устанавливаем тайм-айт для ключа

		_, err = pipe.Exec(h.ctx) // Выполняем команды конвейера
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
	err = h.redisClient.Del(h.ctx, fmt.Sprintf("user:%d", userId)).Err()
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{"ok"})
}

func (h *Handler) deleteItem(c *gin.Context) {
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

	err = h.services.TodoItem.Delete(userId, itemId)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Удаляем все данные из кэша Redis, т.к. у нас нет listId для удаления item:id
	err = h.redisClient.Del(h.ctx, fmt.Sprintf("user:%d", userId)).Err()
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, statusResponse{"ok"})
}
