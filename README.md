# Файлы для итогового задания

В директории `tests` находятся тесты для проверки API, которое должно быть реализовано в веб-сервере.

Директория `web` содержит файлы фронтенда.

# Сборка

docker build --build-arg DATA_DIR=</path/to/db/in/container> --tag my-app:v1 .

# Запуск:

docker run -p <внешний порт>:<TODO_PORT> --volume=</local/path>:<DATA_DIR> -e "TODO_PASSWORD=<password>" -e "TODO_PORT=<port> -e "TODO_DBFILE=<DATA_DIR/db_file_name>" my-app:v1
