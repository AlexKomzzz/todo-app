// Информация по реализации Redis
////////////////////////////////////////////
/*
// При запросах GET данные кэшируются в хэш-таблицу Redis.
//
// Структура хэш-таблицы:
//
// HKEYS: user:'userId' - ключи таблицы;
//
// FIELD: lists (getAllLists), list:'id' (getListById), items:'listId' (getAllItems), item:'id' (getItemById) - поля ключей user:'userId' хэш-таблицы.
// В скобках рядом с полями указаны handler функции GET, которую кэшируют данные в это поле.
//
// Handler функция createList удаляет из хэш-таблицы поле lists.
//
// Handler функции updateList, deleteList, updateItem, deleteItem удаляют из хэш-таблицы весь ключ user:'userId'.
*/

package repository

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type ConfigRedis struct {
	Addr     string
	Password string
	DB       int
}

func NewRedisCache(ctx context.Context, cfg ConfigRedis) (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	status := redisClient.Ping(ctx)

	logrus.Print("Connect status server Redis: ", status)

	redisClient.FlushAll(ctx) // Очистить Redis

	err := status.Err()

	return redisClient, err
}
