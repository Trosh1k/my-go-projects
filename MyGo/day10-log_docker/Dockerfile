# --- 1. Этап сборки (builder) ---
FROM golang:1.22-alpine AS builder

# Устанавливаем рабочую папку внутри контейнера
WORKDIR /app

# Копируем файлы модулей (для кеширования зависимостей)
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем бинарник
RUN go build -o server .

# --- 2. Финальный минимальный образ ---
FROM alpine:latest

# Устанавливаем рабочую папку
WORKDIR /app

# Копируем собранный бинарник из предыдущего этапа
COPY --from=builder /app/server .

# (Опционально) Копируем файл БД, если он уже существует на хосте,
# но лучше не копировать — будем использовать volume
# COPY --from=builder /app/tasks.db* . 2>/dev/null || true

# Открываем порт
EXPOSE 8080

# Команда запуска
CMD ["./server"]