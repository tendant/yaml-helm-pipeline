FROM node:18-alpine AS frontend-builder

WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install
COPY frontend/ ./
RUN npm run build

FROM golang:1.24-alpine AS backend-builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 GOOS=linux go build -o yaml-helm-pipeline ./cmd/server/

FROM alpine:3.18

RUN apk add --no-cache ca-certificates curl helm git

WORKDIR /app

COPY --from=backend-builder /app/yaml-helm-pipeline .
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist
COPY .env.example ./.env.example

ENV PORT=4000
ENV HOST=0.0.0.0

EXPOSE 4000

CMD ["./yaml-helm-pipeline"]
