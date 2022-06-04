# syntax=docker/dockerfile:1

FROM golang:1.18

WORKDIR /todo-app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY ./ ./

RUN go build -o ./bin/main cmd/*.go

EXPOSE 8080

CMD ["./bin/main"]

#migrate -path ./schema -database 'postgres://postgres:qwerty@localhost:5432/postgres?sslmode=disable' up