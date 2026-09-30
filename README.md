# tables

Go library and CLI for [Nextcloud Tables](https://github.com/nextcloud/tables).

## Library

```go
client, err := tables.NewClient(
    "https://cloud.example.net",
    tables.WithAppPassword("stefankuehnel", appPassword),
)
if err != nil {
    return err
}

rows := client.Table(42).Rows()
```

Public-share rows use the same scoped row API:

```go
rows := client.Share(tables.Share{
    Token: tables.ShareToken(token),
    Password: tables.SharePassword(password),
}).Rows()
```

## CLI

The executable is named `tables`.

```text
tables auth login
tables list
tables columns list <table-id>
tables rows list <table-id>
tables rows list --share-token <token>
```

Structured output supports `--json`, `--jq`, and `--template` without invoking an external `jq` binary.

## Development

```bash
nix develop .#devEnvironment
task
```

The default Task workflow runs lint, exact 100% coverage validation, and the Nix build. Integration tests provision disposable real Nextcloud instances with Testcontainers and respect `DOCKER_HOST` for remote Docker daemons.
