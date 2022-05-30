FROM go:1.18

COPY . /var/todo-app

WORKDIR /var/todo-app

EXPOSE 8080

RUN go mod download && go mod verify

RUN go build cmd/*.go

CMD ["./main"]