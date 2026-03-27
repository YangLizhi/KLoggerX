FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o kloggerx-server ./cmd/server

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /app/kloggerx-server .
COPY --from=builder /app/config/config.yaml ./config/
RUN mkdir -p logs
EXPOSE 8178
CMD ["./kloggerx-server"]
