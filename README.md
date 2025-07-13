# yak
CLI for tools maintained by SRE Green

## Install the CLI

Follow [this page](https://doctolib.atlassian.net/wiki/x/iAB_ng).

## Requirements

[Golang](https://go.dev) 1.24

## How to contribute

Check [this](https://doctolib.atlassian.net/l/cp/JSh9KCXg) first.

### How to work on your computer

You can either build the yak/yak-secret cli locally:

```bash
go build ./cmd/yak
go build ./cmd/yak-secret
```

Or directly run the cli with `go run`, e.g.:

```bash
go run cmd/yak/main.go terraform version check
```

### How to test

You can run the whole test suite by running:

```bash
go test ./...
```

To run the tests for a specific package, you can run:

```bash
go test ./internal/cmd/jira
```

### How to release a new version

This project uses Release Please to release new versions. After you have merged a PR with the title prefix `feat` or `fix`, a release please-PR will automatically be created. Once that PR is merged,
the new version is automatically released.

For more information about Release Please, see the official [documentation](https://github.com/googleapis/release-please#release-please).

### Linters

`golangci-lint` is used to lint the code. The configuration file is available in `.golangci.yml`.

To ease development, install `golangci-lint` locally with: `brew install golangci-lint`

In the root folder of this repository, you can now lint the whole codebase with: `golangci-lint run ./...`

If you are using VSCode and official Golang extension, you can change the default linter from `staticcheck` to `golangci-lint` by updating the `Lint Tool` parameter in the extension settings.
