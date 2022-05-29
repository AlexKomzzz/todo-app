# REST API Для Создания TODO Списков на Go  ![GO][go-badge]


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