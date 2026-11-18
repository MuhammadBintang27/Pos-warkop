FROM node:22-alpine AS web
WORKDIR /src
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.26-alpine AS api
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -o /warkop ./cmd/server && CGO_ENABLED=0 go build -o /superadmin ./cmd/superadmin

FROM alpine:3.20
# tzdata: config.Load memanggil time.LoadLocation(TZ_NAME)
RUN apk add --no-cache tzdata ca-certificates
WORKDIR /app
COPY --from=api /warkop /superadmin ./
COPY --from=web /src/dist ./dist
ENV APP_ENV=production PORT=8080 STATIC_DIR=/app/dist TZ_NAME=Asia/Jakarta
EXPOSE 8080
CMD ["./warkop"]
