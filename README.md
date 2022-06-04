# REST API Для Создания TODO Списков на Go v. 1.18


### Prerequisites
- go 1.17
- docker & docker-compose
- [golangci-lint](https://github.com/golangci/golangci-lint) (<i>optional</i>, used to run code checks)
- [swag](https://github.com/swaggo/swag) (<i>optional</i>, used to re-generate swagger documentation)

Create .env file in root directory and add following values:
```dotenv
DB_PASSWORD= <your password>

SOLT=<random string>

JWT_SECRET=<random string>
```

Create config.yml file in root directory and add following values:
```dotenv
*******EXEMPLED********
port: ":8000"

db:
  username: "postgres"
  password: "qwerty"
  host: "localhost"
  port: "5432"
  dbname: "postgres"
  sslmode: "disable"

redis:
  addr: "localhost:6379"
  password: ""
  db: "0"
```

Запустить миграции

  $ migrate -path ./schema -database 'postgres://postgres:qwerty@localhost:5432/postgres?sslmode=disable' up

# Docker

Запустить контейнер Postgres:

  $ docker run --name db -dp 5432:5432 -e POSTGRES_PASSWORD='qwerty' --rm -v roach:/var/lib/postgresql/data --network mynet postgres

Запустить контейнер Redis:

  $ docker run --name redis --rm --network mynet -dp 6379:6379 redis

Запустить контейнер API:

  $ docker run -it -dp 8080:8080 --network mynet --name apidb -e DB_PASSWORD='qwerty' api