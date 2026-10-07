# Dosage Drugs — веб-сервис

## Технологии
- Go, Gin, GORM, PostgreSQL, MinIO
- API для SPA-приложения «Дозирование лекарств для детей»

## Стек сервисов (docker-compose)
| Сервис | Порт | Назначение |
|---|---|---|
| PostgreSQL | 5435 | БД |
| MinIO API | 9000 | Хранилище файлов |
| MinIO Console | 9001 | Веб-консоль MinIO |
| Adminer | 8081 | Веб-клиент БД |
| Go-сервер | 8090 | API и HTML-страницы |

## Таблицы БД

### users
| Поле | Тип | Описание |
|---|---|---|
| id | int | PK |
| login | varchar(25) | уникальный логин |
| password | varchar(100) | пароль (не отдаётся в JSON) |
| is_moderator | bool | флаг модератора |

### drugs
| Поле | Тип | Описание |
|---|---|---|
| id | int | PK |
| title | varchar(100) | название препарата |
| description | text | описание |
| status | varchar(20) | черновик / опубликован / Удален |
| image_key | varchar(255) | имя файла-картинки в MinIO |
| video_key | varchar(255) | имя файла-видео в MinIO |
| adult_dose | int | взрослая доза, мг/кг |
| concentration | float | концентрация, % |
| creator_id | int | FK → users.id |
| created_at | timestamp | дата создания |
| formed_at | timestamp | дата публикации |

### likes
| Поле | Тип | Описание |
|---|---|---|
| id | int | PK |
| drug_id | int | FK → drugs.id |
| user_id | int | FK → users.id |

## HTTP API (все с префиксом `/api`)

### Домен «услуга»

| Метод | URL | Описание | Тело / параметры |
|---|---|---|---|
| GET | /api/drugs | Список опубликованных с фильтром | `?search=Нуро` |
| GET | /api/drugs/feed | Лента (без id) | — |
| GET | /api/drugs/draft | Черновик текущего пользователя | — |
| POST | /api/drugs | Создание (multipart) | `title`, `image`, `video` |
| PUT | /api/drugs/:id/publish | Публикация | JSON: `title`, `adult_dose`, `concentration_percent`, `description` |
| DELETE | /api/drugs/:id | Soft delete | — |
| POST | /api/drugs/:id/like | Лайк 0/1 | JSON: `{"value": 1}` |

### Домен «пользователь»

| Метод | URL | Описание | Тело |
|---|---|---|---|
| POST | /api/users/register | Регистрация | `{"login":"...","password":"..."}` |
| POST | /api/users/login | Заглушка логина | — |
| POST | /api/users/logout | Заглушка логаута | — |

## Формат ответов

Успех:
```json
{"status": "success", "data": {...}}
