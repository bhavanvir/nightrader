FROM golang:1.21

WORKDIR /app

# Copy local code to the container image.
COPY . .

# Build the command inside the container.
RUN go build -v -o app .

# Run the service binary.
CMD ["./app"]