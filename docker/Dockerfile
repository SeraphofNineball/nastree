# ---- frontend build ----
FROM node:22-alpine AS frontend
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# ---- backend build ----
FROM golang:1.23-alpine AS backend
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /nastree ./cmd/server

# ---- runtime ----
FROM alpine:3.20
RUN apk add --no-cache tzdata
WORKDIR /app
COPY --from=backend /nastree ./nastree
COPY --from=frontend /app/dist ./static

ENV SCAN_PATH=/data \
    DB_PATH=/db/nastree.db \
    STATIC_DIR=/app/static \
    LISTEN_ADDR=:8080 \
    SCAN_INTERVAL=6h \
    RETAIN_SCANS=5

EXPOSE 8080
VOLUME ["/db"]
ENTRYPOINT ["./nastree"]
