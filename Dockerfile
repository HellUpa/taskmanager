# Базовый образ для сборки (содержит Go).
FROM golang:1.23.6 AS builder

# Устанавливаем рабочую директорию внутри контейнера.
WORKDIR /app

# Копируем файлы go.mod и go.sum.
COPY go.mod go.sum ./

# Скачиваем зависимости.  Это кэшируется Docker'ом,
# и зависимости будут перекачиваться только если go.mod/go.sum изменились.
RUN go mod download

# Копируем исходный код.
COPY . .

# Компилируем приложение.
RUN CGO_ENABLED=0 GOOS=linux go build -o taskmanager-server ./cmd/server

# ----------------------------------------------------
# Финальный образ (минимальный, только для запуска).
FROM alpine:latest

# Создаём директорию для приложения.
WORKDIR /app

# Копируем скомпилированный бинарник из *builder* образа.
COPY --from=builder /app/taskmanager-server .

# Указываем команду, которая будет запускаться при старте контейнера.
# Здесь же указываем параметры запуска.
CMD ["./taskmanager-server", "-host", "db", "-port", "5432", "-user", "postgres", "-password", "postgres", "-dbname", "taskmanager", "-listen", "50051", "-healthcheck", "8080"]