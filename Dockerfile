# Stage 1: Build frontend
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Stage 2: Build backend
FROM golang:1.25-alpine AS backend
WORKDIR /app/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -o /pauls-bach .

# Stage 3: Run
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend /pauls-bach ./pauls-bach
COPY --from=frontend /app/frontend/dist ./frontend/dist

ENV PORT=8080
ENV DATA_DIR=/app/data
ENV FRONTEND_DIST=/app/frontend/dist

RUN mkdir -p /app/data
VOLUME /app/data

EXPOSE 8080
CMD ["./pauls-bach"]
