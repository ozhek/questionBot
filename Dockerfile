# syntax=docker/dockerfile:1.4

# Use the official Golang image to build the application
FROM --platform=$BUILDPLATFORM golang:1.22 as builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the Go app for Linux AMD64
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o qaBot ./cmd/main.go

# Final minimal image
FROM scratch

WORKDIR /app

COPY --from=builder /app/qaBot .

CMD ["/app/qaBot"]