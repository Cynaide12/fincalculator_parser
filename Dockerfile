FROM golang:1.23.4-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
##изменить название бинарника (вместо arch)
RUN CGO_ENABLED=0 GOOS=linux go build -o fincalparser ./cmd/main.go

FROM alpine:latest
RUN apk add --no-cache
WORKDIR /app
COPY --from=builder /src .
EXPOSE 8080
## изменить на папку, в которую будет ложиться бэк в контейнере
CMD ["./app"] 
