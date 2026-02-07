FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o smartlattes ./cmd/smartlattes

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /build/smartlattes /smartlattes
EXPOSE 8080
CMD ["/smartlattes"]
