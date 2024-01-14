# Gopher Mart

Проекта курса «Go-разработчик»



# Curl запросы для быстрого тестирования хендлеров

## GopherAccural

### Order info 
```bash
curl -v -H "Content-Type: text/plain" http://localhost:8080/api/orders/8327568377
```
### Add order
```bash
curl -v -H "Content-Type: application/json" -X POST http://localhost:8080/api/orders -d '{"order":"8327568377","goods":[{"description":"Чайник Bork","price":7000}]}'
```


curl -v -H "Content-Type: application/json" -X POST http://localhost:8080/api/shorten -d '{"url":"https://practicum1.yandex1.ru"}'

# Работа с шаблоном проекта

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без
   префикса `https://`) для создания модуля

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m master template https://github.com/yandex-praktikum/go-musthave-diploma-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/master .github
```

Затем добавьте полученные изменения в свой репозиторий.
