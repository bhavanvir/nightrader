FROM golang:1.21

WORKDIR /app

COPY . .

RUN go get github.com/lib/pq

RUN go get github.com/google/uuid

RUN go get github.com/rabbitmq/amqp091-go

RUN go build -v -o app .

CMD ["./app"]