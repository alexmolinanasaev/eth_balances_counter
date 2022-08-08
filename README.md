# Как запустить?

## Локально

```sh
export API_KEY=<YOUR_API_KEY>
export NODE_URL=<NODE_HTTP_ENDPOINT>
go run .
```

## С помощью docker-compose

1. Создать файл с названием **.env** в корневой папке проекта и заполнить согласно файлу .env.example
2. Запустить команду ```sh docker-compose up --build -d```

Чтобы посмотреть логи при запуске введите команду ```sh docker logs -f eth_balances_counter```
