# lan-share

[![CI](https://github.com/gertyhiler/lan-share/actions/workflows/ci.yml/badge.svg)](https://github.com/gertyhiler/lan-share/actions/workflows/ci.yml)

A small browser-based chat for sharing text and files between devices on the same
local network. I use it daily to move things between my work and personal laptops.
Run one Go binary, open its LAN address on another device, and share. No internet
connection or client installation is needed at runtime.

Designed for a **trusted LAN**, not the public internet. Read [SECURITY.md](SECURITY.md)
for the access boundary. The current browser interface is in Russian.

## Install and run

Requires Go 1.22 or newer to install from source:

```sh
go install github.com/gertyhiler/lan-share/cmd/lanshare@latest
```

The binary is installed in `$GOPATH/bin`, or `$(go env GOPATH)/bin` by default.
Make sure that directory is on your PATH. Choose where local chat and files live:

```sh
mkdir -p ~/lan-share-data
lanshare --host 0.0.0.0 --port 8000 --root ~/lan-share-data
```

Open the LAN URL printed by the server on each device, for example
`http://192.168.1.10:8000/`. Devices must be on a network that permits them to reach
the host. If it does not open, check the host firewall and network isolation.
Stop the server with Ctrl-C.

For a local-only development session, bind to `127.0.0.1` instead. `--root` defaults
to the current working directory; the default listener is `0.0.0.0:8000`.

## What it does

- One shared chat with live messages and participants through SSE.
- Text, file attachments, and inline previews for image/video attachments.
- A limited Markdown subset: links, bold, italic, inline code and fenced code.
  Message HTML is not executed; inline code can be copied by clicking it.
- Shared-file and paste endpoints for small local scripts.

The server associates devices with their direct connection IP and sets an
HttpOnly cookie. Names are generated from the server-side device ID. This is a
convenience identity for the LAN chat, not user authentication. IP changes or
shared addresses can affect identity.

## Local data

The server creates these directories under `--root`:

| Directory | Contents |
| --- | --- |
| `lan_share_uploads/` | Uploaded files and chat attachments |
| `lan_share_shared/` | Files placed here for sharing on the LAN |
| `lan_share_pastes/` | Saved pastes, including `latest.txt` |
| `lan_share_chat/` | Chat history and device/IP mapping |

These are runtime data, not source files. Keep them out of Git and choose a
storage directory appropriate for the material you share.

## HTTP integration

- `GET /api/chat/stream` — SSE events: `history`, `message`, `participants`.
- `POST /api/chat/messages` — JSON with `text` and `attachments`.
- `POST /upload` with `Accept: application/json` — attachment upload, returning
  `{"ok": true, "files": [...]}`.
- `POST /paste`, `GET /api/paste/latest` — legacy paste API retained for scripts.

## Development

```sh
git clone https://github.com/gertyhiler/lan-share.git
cd lan-share
go run ./cmd/lanshare --host 127.0.0.1 --port 8000
go vet ./...
go test ./... -race
```

Build a binary with `go build -o lanshare ./cmd/lanshare`.

The Go implementation separates domain contracts, use cases and adapters:

| Responsibility | Location |
| --- | --- |
| Domain entities and storage contracts | `internal/domain` |
| Chat, file and paste operations | `internal/usecase` |
| Filesystem storage and HTTP interface | `internal/adapter` |
| Configuration and application wiring | `cmd/lanshare` |

See [CONTRIBUTING.md](CONTRIBUTING.md) for focused changes and fork instructions.
[MIT licensed](LICENSE).
