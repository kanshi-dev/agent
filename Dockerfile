FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY . .

RUN go build -o kanshi-agent cmd/core/main.go


FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/kanshi-agent .

CMD ["./kanshi-agent"]