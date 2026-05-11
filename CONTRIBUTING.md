# Contributing

Thanks for your interest in improving lan-share.

## Development

- Go 1.22 or newer (see `go.mod`).
- From the repository root:

```bash
go vet ./...
go test ./... -race
```

- Run the server locally:

```bash
go run ./cmd/lanshare --host 127.0.0.1 --port 8000
```

## Pull requests

- Keep changes focused on a single concern.
- Run `gofmt` on edited Go files (`gofmt -w .`).
- Add or update tests when changing behavior.

## Module path

The module is published as `github.com/gertyhiler/lan-share`. If you maintain a fork under another path, update the `module` line in `go.mod` and replace import prefixes consistently (for example with your editor or `sed`).

## Runtime data

Directories `lan_share_uploads/`, `lan_share_shared/`, and `lan_share_pastes/` are local data created at runtime. Do not commit their contents; they are listed in `.gitignore`. If they were ever committed by mistake, remove them from Git history before going public.
