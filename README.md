# File exporter

Prometheus экспортер для мониторинга изменения состояния директории написанный на Golang.

## Установка

```file_exporter``` по стандарту слушает ```:9393``` HTTP-порт. Смотрите ```-h, --help``` для большей информации

### Docker

Для сборки:

```
git clone https://github.com/IvanSaratov/file_exporter.git
cd file_exporter
docker build -t file_exporter .
```

По стандарту экспортер наблюдает за изменениям в той же папке в которой запущен, 
для изменения надо использовать флаг ```-d, --directory```
```
docker run -p 9393:9393 file_exporter -d /path/for/watch
```

Так же если хотите мониторить папку в хост-системе, 
то её придется монитировать к контейнеру при запуске и указывать ключом для экспортера за чем наблюдать:
```
docker run -v /var/lib/interesting_dir:/path/for/watch -p 9393:9393 file_exporter -d /path/for/watch
```

### Локальный запуск

Требования:
* Go компилятор

Сборка:
```
git clone https://github.com/IvanSaratov/file_exporter.git
cd file_exporter
go build -o file_exporter .
```

Что бы увидить все доступные команды:
```
./file_exporter -h
```

## Метрики

Имя | Описание
----|---------
file_exporter_directory_get_size | Размер папки в байтах
file_exporter_directory_get_last_update | Время последнего события в секундах по UNIX времени (В тегах описанние самого события, например {status="WRITE"})
