# tables

`tables` is a Go library and command-line client for [Nextcloud Tables](https://github.com/nextcloud/tables). It supports authenticated Tables access, public link shares, table/column/row CRUD, multiple accounts, secure local credential storage, and GitHub CLI-style structured output.

## Go library

```go
package main

import (
    "context"
    "log"

    "github.com/stefankuehnelai/tables/pkg/tables"
)

func main() {
    client, err := tables.NewClient(
        "https://cloud.example.net",
        tables.WithAppPassword("stefankuehnel", "app-password"),
    )
    if err != nil {
        log.Fatal(err)
    }

    rows, err := client.Table(42).Rows().List(
        context.Background(),
        tables.ListRowsOptions{Limit: 100},
    )
    if err != nil {
        log.Fatal(err)
    }
    _ = rows
}
```

Public shares use the same row API:

```go
rows := client.Share(tables.Share{
    Token:    tables.ShareToken(token),
    Password: tables.SharePassword(password),
}).Rows()
```

Both table and share scopes expose `List`, `Create`, `Update`, and `Delete` row operations. Password-protected share sessions use isolated in-memory cookie jars and are never persisted by the library.

## CLI

The binary is named `tables`.

Log in with a password from stdin:

```bash
printf '%s' "$PASSWORD" | tables auth login \
  --hostname cloud.example.net \
  --username stefankuehnel \
  --password-stdin
```

Or use an app password:

```bash
printf '%s' "$APP_PASSWORD" | tables auth login \
  --hostname cloud.example.net \
  --username stefankuehnel \
  --app-password-stdin
```

Interactive credentials are stored in the operating-system keyring. Non-secret account metadata is stored in the user config directory. There is no plaintext-secret fallback.

Multiple hosts and users are supported:

```bash
tables auth status
tables auth switch --hostname cloud.example.net --username johndoe
tables auth logout --hostname cloud.example.net --username johndoe
```

Core commands include:

```text
tables list
tables get <table-id>
tables create --title <title>
tables update <table-id>
tables delete <table-id>

tables columns list <table-id>
tables columns create <table-id> --title <title> --type <type>
tables columns update <table-id> <column-id>
tables columns delete <table-id> <column-id>

tables rows list <table-id>
tables rows create <table-id> --value <column-id>=<value>
tables rows update <table-id> <row-id> --value <column-id>=<value>
tables rows delete <table-id> <row-id>
```

Public shares do not require a table ID:

```bash
tables rows list --server https://cloud.example.net --share-token "$TOKEN"
tables rows list --server https://cloud.example.net \
  --share-token "$TOKEN" \
  --share-password "$SHARE_PASSWORD"
```

### Structured output

```bash
tables list --json id,title
tables list --json id,title --jq '.[] | select(.title == "Customers")'
tables list --json id,title --template '{{range .}}{{.id}} {{.title}}{{"\n"}}{{end}}'
```

`--jq` uses a pure-Go implementation; no external `jq` binary is required.

### Environment variables

CLI flags take precedence over environment variables, which take precedence over the active stored account. Supported variables include:

```text
NEXTCLOUD_TABLES_HOSTNAME
NEXTCLOUD_TABLES_SERVER
NEXTCLOUD_TABLES_USERNAME
NEXTCLOUD_TABLES_PASSWORD
NEXTCLOUD_TABLES_APP_PASSWORD
NEXTCLOUD_TABLES_SHARE_TOKEN
NEXTCLOUD_TABLES_SHARE_PASSWORD
```

This makes headless CI use possible without a desktop keyring.

## Development

The repository follows the Nix/Task workflow conventions of `stefankuehnel/calculator`.

Enter the development environment:

```bash
nix develop .#devEnvironment
```

Run the default workflow:

```bash
task
```

The default task runs, in order:

1. `lint`
2. `test:coverage`
3. `build`

Tests use Ginkgo + Gomega. The coverage gate is exactly `100.0%` with `--coverpkg=./...`. Integration tests provision a real pinned Nextcloud + Tables instance with Testcontainers and run in the normal test workflow. Set `DOCKER_HOST` when using a remote Docker daemon.

Compatibility lanes can override the pinned integration baseline with:

```text
NEXTCLOUD_TABLES_TEST_NEXTCLOUD_IMAGE
NEXTCLOUD_TABLES_TEST_TABLES_VERSION
```

Run the CI workflow locally inside the Nix CI environment with:

```bash
task run:workflow:local WORKFLOW_NAME=ci
```

## License

GPL-3.0. See [LICENSE](LICENSE).
