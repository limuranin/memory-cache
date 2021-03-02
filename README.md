# memory-cache
Простая имплементация кеша в памяти

## Описание проекта
Проект включает в себя: 
* имплементацию кеша в памяти
* сервер для запуска кеша
* API клиент к серверу
* unit тесты кеша
* интеграционные тесты с обращением к кеш-серверу через API клиент

Кеш и API клиент к кеш-серверу реализуют общий интерфейс Cacher.
Это дает возможность перейти с локального кеша на удаленный без изменения кода.

```
type Cacher interface {
    Set(key string, value interface{}, ttl time.Duration) error
    Get(key string) (interface{}, error)
    GetListElem(key string, index int) (interface{}, error)
    GetMapElemValue(key string, mapKey string) (interface{}, error)
    Remove(key string) error
    Keys() ([]string, error)
}
```

## Сборка и запуск
* Запускаем команду: `make build`

На выходе в директории cmd/memory-cache/ появится файл memory-cache

## Запуск unit тестов
* Запускаем команду: `make test`

## Запуск интеграционных тестов
* Запускаем команду: `make test_integration`

Интеграционные тесты являются примером использования memory-cache сервера и клиента.
Они запускают memory-cache сервер, с помощью клиента отправляют набор базовых запросов и проверяют результат

Более полное тестирование кеша произведено в unit тестах

## Конфигурация
Приложение конфигурируется через environment переменные:

| KEY |TYPE | DEFAULT | DESCRIPTION   | 
|---|---|---|---|
| MC_SERVER_LISTEN_ADDRESS  | String  | 127.0.0.1:8080  | Server listen address   | 
| MC_CACHE_CLEANING_INTERVAL  | Duration  | 30s  | Cleaning cache interval   |

## Документация
Спецификация к клиенту находится в файле swagger.yml


