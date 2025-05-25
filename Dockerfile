# Use the official Golang image as the base image
FROM golang:1.22 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the application as a statically linked binary
RUN CGO_ENABLED=0 go build -o ./qaBot ./cmd/main.go

# Use scratch as the base image for the final container
FROM scratch

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/qaBot .

# Command to run the application
CMD ["/app/qaBot", "-config=test"]