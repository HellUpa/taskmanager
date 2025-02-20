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
go run cmd/server/main.go
```
## Поддерживает следующие параметры:
>-host: имя хоста или IP-адрес вашего сервера PostgreSQL. Значение по умолчанию — localhost.  
>-port: порт, на котором PostgreSQL слушает. Значение по умолчанию — 5432.  
>-user: имя пользователя PostgreSQL. Значение по умолчанию — postgres.  
>-password: пароль PostgreSQL. Значение по умолчанию — postgres.  
>-db_name: имя базы данных PostgreSQL. Значение по умолчанию — postgres.  

## Запуск клиента (для тестирования)
На данный момент просто тестирует операции с БД с уже зашитыми данными.
```
go run cmd/client/main.go
```
