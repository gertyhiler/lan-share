# lan-share

[![CI](https://github.com/gertyhiler/lan-share/actions/workflows/ci.yml/badge.svg)](https://github.com/gertyhiler/lan-share/actions/workflows/ci.yml)

Минимальный офлайн-сервис для локалки: общий чат устройств и передача файлов между ними через браузер (без интернета). Реализация на **Go** (слои: `domain` → `usecase` → `adapter`).

Сервис рассчитан на **доверенную LAN**, не на публикацию в интернет. См. [SECURITY.md](SECURITY.md).

## Установка

Требуется Go 1.22+.

```bash
go install github.com/gertyhiler/lan-share/cmd/lanshare@latest
```

После установки бинарник `lanshare` окажется в `$GOPATH/bin` (или в `$(go env GOPATH)/bin`). При другом модуле форка замените путь в команде и в `go.mod` — см. [CONTRIBUTING.md](CONTRIBUTING.md).

## Запуск

Из корня репозитория (чтобы каталоги `lan_share_*` создались рядом с проектом):

```bash
go run ./cmd/lanshare --host 0.0.0.0 --port 8000
```

Сборка бинарника:

```bash
go build -o lanshare ./cmd/lanshare
./lanshare --host 0.0.0.0 --port 8000
```

Флаг `--root` задаёт каталог, в котором создаются `lan_share_uploads`, `lan_share_shared`, `lan_share_pastes`, `lan_share_chat` (по умолчанию текущая рабочая директория).

Открой на другом устройстве в той же сети:

- `http://<LAN-IP>:8000/`

## Слои проекта

| Слой     | Пакет                                          | Назначение                   |
| -------- | ---------------------------------------------- | ---------------------------- |
| Domain   | `internal/domain`                              | сущности, контракты хранилищ |
| Use case | `internal/usecase/...`                         | сценарии: чат, паста, файлы  |
| Adapters | `internal/adapter/fs`, `internal/adapter/http` | диск и HTTP                  |
| Вход     | `cmd/lanshare`                                 | флаги, DI, HTTP-сервер       |

## Что куда кладётся

- Файлы, загруженные с других устройств → `lan_share_uploads/`
- Файлы для раздачи в LAN → `lan_share_shared/`
- История чата и привязка LAN IP → device id → `lan_share_chat/`
- Текстовые пасты → `lan_share_pastes/` (последняя версия ещё в `lan_share_pastes/latest.txt`)

Эти каталоги создаются при работе и **не должны коммититься** (см. `.gitignore`).

## Чат

Главная страница теперь работает как один общий чат. Каждое устройство получает серверный `deviceId`: сервер нормализует IP из прямого `RemoteAddr`, сохраняет привязку в `lan_share_chat/devices.json` и выставляет `HttpOnly` cookie `lan_share_device`. JavaScript не читает и не отправляет `deviceId`; имя устройства используется только как отображаемая подпись.

- `GET /api/chat/stream` — SSE-события `history`, `message`, `participants`
- `POST /api/chat/messages` — JSON-сообщение с `text`, `displayName`, `attachments`
- `POST /upload` с `Accept: application/json` — загрузка файлов для вложений, ответ `{"ok": true, "files": [...]}`

Legacy API пасты оставлен для скриптов: `POST /paste`, `GET /api/paste/latest`.

## Полезные команды

Узнать LAN IP на macOS (часто Wi‑Fi — `en0`):

```bash
ipconfig getifaddr en0
```

Если клиент не открывает страницу — проверьте, что фаервол не блокирует входящие на выбранный порт.

## Участие и лицензия

- [CONTRIBUTING.md](CONTRIBUTING.md) — как собирать, тестировать и слать PR.
- [LICENSE](LICENSE) — MIT.
- Уязвимости: [SECURITY.md](SECURITY.md).

Форк: замените `module` в `go.mod` и префикс импортов на путь вашего репозитория, обновите бейдж CI и ссылку в [SECURITY.md](SECURITY.md).
