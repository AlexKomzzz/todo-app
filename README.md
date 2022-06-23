# REST API Для Создания TODO Списков на Go v. 1.18
Status of Last Deployment:<br>
<img src="https://github.com/AlexKomzzz/todo-app/workflows/CI-CD_todo-app/badge.svg?branch=master"><br>

### Предпосылки
- go 1.18 
- [Gin](https://github.com/gin-gonic/gin) (веб-фреймворк)
- PostgreSQL (БД: [sqlx](https://github.com/jmoiron/sqlx) и [migrate](https://github.com/golang-migrate/migrate))
- [Redis](https://github.com/go-redis/redis) (кэширование данных)
- docker & docker-compose
- GitHub Action (состоит из трех шагов: тестирование, коннект с удаленным хостом и деплой)
- [swag](https://github.com/swaggo/swag) (<i>optional</i>, used to re-generate swagger documentation)
- Unit-тестирование (исп. [gomock](https://github.com/golang/mock), 
[go-sqlmock](https://github.com/DATA-DOG/go-sqlmock), [redismock](https://github.com/go-redis/redismock))
- Использование HTML-template для ошибки 404 при несуществующем URL (с применением [go:embed](https://pkg.go.dev/embed))
- Использование [JWT](https://github.com/golang-jwt/jwt) для аутентификации и авторизации

## Start use

Создайте файл .env в корневом каталоге со следующим значением:
```dotenv
DB_PASSWORD= <your password>
```

Создайте файл config.yml в корневом каталоге и добавьте следующие значения:
```dotenv
*******ПРИМЕР********
port: ":8000"

db:
  username: "postgres"
  password: "qwerty"
  host: "localhost" or "db"
  port: "5432"
  dbname: "postgres"
  sslmode: "disable"

redis:
  addr: "localhost:6379" or "redis:6379"
  password: ""
  db: "0"
```


# Docker

Создать образ из Dockerfile.multi:

    $ docker build -f Dockerfile.multi -t api .

Запустить контейнер Postgres:

    $ docker run --name db -dp 5432:5432 -e POSTGRES_PASSWORD='qwerty' --rm -v roach:/var/lib/postgresql/data --network mynet postgres

Запустить контейнер Redis:

    $ docker run --name redis --rm --network mynet -dp 6379:6379 redis

Запустить контейнер API:

    $ docker run -it -dp 8000:8000 --network mynet --rm --name apidb -e DB_PASSWORD='qwerty' api

# Docker compose
Запустить приложение одной командой:

    $ docker compose up --build -d



# Миграции
Использовать миграции для создания таблиц в БД (применить свои настройки подключения к БД):

    $ migrate -path ./schema -database 'postgres://postgres:qwerty@localhost:5432/postgres?sslmode=disable' up


#  Использование swagger
После запуска приложения перейдите по ссылке:
 http://localhost:8000/swagger/index.html#/




#
#
 Выполнено на основании: https://github.com/zhashkevych/todo-app.git

