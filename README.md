## Описание
Этот проект - простейший менеджер задач.
## Настройка БД
```sql
    CREATE TABLE tasks (
        id SERIAL PRIMARY KEY,
        title VARCHAR(255) NOT NULL,
        description TEXT,
        due_date TIMESTAMP,
        completed BOOLEAN DEFAULT FALSE,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
```
## Запуск сервера
Команда запуска сервера, из корневой папки:
```
docker-compose up --build
```

## Запуск клиента (для тестирования)
На данный момент просто тестирует операции с БД с уже зашитыми данными.
```
go run cmd/client/main.go
```
