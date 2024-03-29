# Gopher Mart

Проекта курса «Go-разработчик»


[Диаграмма классов проекта](https://miro.com/app/board/uXjVNx6fYjE=/?share_link_id=94199426221)

---

# Настройка среды проекта

Для работы приложения Gophermart необходимо в контейнере запустить базу данных Postgres и выполнить запуск приложения для расчета вознаграждения Accruals (папка проекта cmd/accruals)

## Переменные окружения и флаги

Переменная окружения DSN:
```bash
export DATABASE_URI=postgresql://market:1@localhost/market
```
Переменная RUN_ADDRESS - адреc и порт сервиса Gophermart:
```bash
export RUN_ADDRESS=localhost:8088
```
Адрес и порт системы вознаграждения Accural:
```bash
export ACCRUAL_SYSTEM_ADDRESS=localhost:8090
```

Флаги:
```txt
-a    адреc и порт сервиса Gophermart
-r    адреc и порт системы вознаграждения Accural
-d    DSN подключения базы данных (Data Source Name)
-p    Секрет для шифрования токена JWT
```
## Запуск Postgres в контейнере

Для запуска и остановки Postgres в контейнере выполнятьются скрипты создания и миграции базы в make-файле:
* Инициализация
```bash
make pg
```
* Миграция goose
```bash
https://github.com/pressly/goose
GOOSE_DRIVER=postgres
GOOSE_DBSTRING="postgresql://market:1@localhost/market"
GOOSE_DRIVER=postgres GOOSE_DBSTRING="postgresql://market:1@localhost/market" goose up
GOOSE_DRIVER=postgres GOOSE_DBSTRING="postgresql://postgres:postgres@postgres/praktikum" goose -dir ./migrations  up
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="postgresql://postgres:postgres@postgres/praktikum"
```
```bash
make pg-up
```
* Остановка и удаление контейнера
```bash
make pg-stop
```


## Закуск сервиса рачета вознаграждения

```bash
./accrual_linux_amd64 -a localhost:8090 -d 'postgresql://bonus:1@postgres/bonus?sslmode=disable'
```


# Curl запросы для быстрого тестирования хендлеров

## GopherAccrual

### Order info 
```bash
curl -v -H "Content-Type: text/plain" http://localhost:8080/api/orders/8327568377
```
### Add order
```bash
curl -v -H "Content-Type: application/json" -X POST http://localhost:8080/api/orders -d '{"order":"8327568377","goods":[{"description":"Чайник Bork","price":7000}]}'


curl -v -H "Content-Type: application/json" -X POST http://localhost:8080/api/orders -d '{"order":"5536373433","goods":[{"description":"Колпак Я люблю баню войлок б40273","price":143},{"description":"Штора д/бережливых 1065BL 170*180см","price":295}]}'
```

# GopherMartUser - клиент для примера работы с приложением GopherMart и его тестирования

Простой клиент для проверки и работы с приложением:
[GopherMartUser](https://github.com/shulganew/gophermartuser) 

Флаги:
```txt
-d    DSN подключения базы данных (Data Source Name)
-m    Создание запросов к системе GopherMart:
		-m 1 - Register new user
		-m 2 - Login user
		-m 3 - Get users order list
		-m 4 - Set user order
		-m 5 - Get user's balanse
		-m 6 - Make withdrawn
		-m 7 - Get user's withdrawals
```

## Mock generate 

```bash
go install github.com/golang/mock/mockgen@v1.6.0
go get github.com/golang/mock/gomock

```

```bash
mockgen -source=internal/services/calc.go \
    -destination=internal/services/mocks/calc_mock.gen.go \
    -package=mocks

mockgen -source=internal/services/accrual.go \
    -destination=internal/services/mocks/accrual_mock.gen.go \
    -package=mocks

mockgen -source=internal/services/orders.go \
    -destination=internal/services/mocks/orders_mock.gen.go \
    -package=mocks

mockgen -source=internal/services/user.go \
    -destination=internal/services/mocks/user_mock.gen.go \
    -package=mocks
```

