# Multi-stage build for a minimal final image
FROM golang:1.23-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o shinemonitor-mqtt-bridge main.go

# --- Final image ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /build/shinemonitor-mqtt-bridge .

EXPOSE 8080
ENTRYPOINT ["./shinemonitor-mqtt-bridge"]
