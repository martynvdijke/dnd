FROM node:24-alpine AS ts-builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY tsconfig.json vite.config.ts ./
COPY ts/ ts/
RUN npm run build:ts

FROM golang:1.26.4-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ts-builder /app/static/js ./static/js
RUN CGO_ENABLED=0 GOOS=linux go build -tags sqlite_fts5 -o villum-server .

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
ENV DOCKER=true
COPY --from=builder /app/villum-server .
COPY --from=builder /app/static ./static
RUN mkdir -p /db /app/media /app/backups && chmod 777 /db /app/media /app/backups
EXPOSE 6280
CMD ["./villum-server"]
