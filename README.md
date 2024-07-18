# go-musthave-shortener-tpl

Проект «Сервис сокращения URL».

## Реализованы следующие функции

1. Организовано три альтернативных типа хранилища:
   - мапа (потокобезопасная);
   - файл JSON;
   - база данных PostgreSQL.
2. Основные хендлеры покрыты юнит-тестами, в тч mock.
3. Реализованы middleware:
   - со сжатием и обратно по методу gzip;
   - с логированием длительности выполнения запроса;
   - с аутентификацией посредством JWT и cookie.
4. Реализована пакетная обработка данных из json


## Предстоит решить

1. Паттерны многопоточности (асинхронные операции с БД)
2. Авторизация
3. Деплой проекта (GitHub Action)

