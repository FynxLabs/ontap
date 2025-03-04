FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY api/ .
RUN go mod download
RUN go build -o api-server .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/api-server .
EXPOSE 8080
CMD ["./api-server"]
