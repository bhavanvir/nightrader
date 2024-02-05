# Use the official Golang image to create a build artifact.
FROM golang:1.21

# Set the working directory inside the container.
WORKDIR /go/app

# Copy local code to the container image.
COPY . .

# Fetch the required package.
RUN go get github.com/lib/pq

# Build the command inside the container.
RUN go build -v -o app .

# Run the service binary.
CMD ["./app"]
