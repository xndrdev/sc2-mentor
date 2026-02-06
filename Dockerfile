# Build Frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
ARG VITE_UMAMI_SCRIPT_URL
ARG VITE_UMAMI_WEBSITE_ID
ENV VITE_UMAMI_SCRIPT_URL=${VITE_UMAMI_SCRIPT_URL}
ENV VITE_UMAMI_WEBSITE_ID=${VITE_UMAMI_WEBSITE_ID}
RUN npm run build

# Build Backend
FROM golang:1.22-alpine AS backend-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o server ./cmd/server

# Production Image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Backend binary
COPY --from=backend-builder /app/server .

# Frontend static files
COPY --from=frontend-builder /app/frontend/dist ./static

# Create data directory
RUN mkdir -p /app/data/uploads

# Non-root user for security
RUN adduser -D -u 1000 appuser
RUN chown -R appuser:appuser /app
USER appuser

ENV GO_ENV=production
ENV PORT=3000

EXPOSE 3000

CMD ["./server", "-port", "3000", "-db", "/app/data/sc2analytics.db", "-uploads", "/app/data/uploads"]
