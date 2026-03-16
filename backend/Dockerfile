FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /api ./cmd/api
RUN CGO_ENABLED=0 go build -o /scraper ./cmd/scraper

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /api /app/api
COPY --from=builder /scraper /app/scraper
RUN mkdir -p /app/data
EXPOSE 8080
CMD ["/app/api"]
