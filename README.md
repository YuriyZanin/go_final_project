# Файлы для итогового задания

В директории `tests` находятся тесты для проверки API, которое должно быть реализовано в веб-сервере.

Директория `web` содержит файлы фронтенда.

# Для сборки 

docker build --tag my-app:v1 .

# Запуск с параметрами по умолчанию

docker run -p <внешний порт>:<TODO_PORT> my-app:v1

# Запуск с параметрами окружения:

docker run -p <внешний порт>:<TODO_PORT> -e "TODO_PASSWORD=<password>" -e "TODO_PORT=<port> -e "TODO_DBFILE=<path>" my-app:v1
