# syntax=docker/dockerfile:1.4

# Build stage with CGO enabled and required dev libs
FROM --platform=$BUILDPLATFORM debian:bullseye as builder

RUN apt-get update && \
    apt-get install -y golang gcc libc6-dev libsqlite3-dev ca-certificates && \
    update-ca-certificates

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -o qaBot ./cmd/main.go

# Final minimal image
FROM scratch

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/qaBot .

CMD ["/app/qaBot"]