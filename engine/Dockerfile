FROM golang:1.21

WORKDIR /app

COPY . .

RUN go get github.com/lib/pq

RUN go build -v -o app .

CMD ["./app"]