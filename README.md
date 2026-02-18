# go-devtools scaffold

This project is a scaffold for a terminal developer-tools CLI in Go.

## What it includes

- Arrow-key menu navigation (`up/down`, `enter`, `left`, `q`)
- Standalone tool modules with a shared interface
- Nested submenus via a common menu builder
- Reusable exit actions (`WithBack`, `WithQuit`)
- Optional per-tool requirement checks (command/env prechecks)
- Install action for missing requirements (`i` key when available)
- Maximum menu nesting depth set to 4

## Run

```bash
go run ./cmd/devtools
```

## GitHub build artifacts and releases

This repo includes a GitHub Actions workflow at `.github/workflows/build-release.yml`.

- On every pull request and push to `main`, it builds Linux/macOS binaries and uploads them as workflow artifacts.
- On tags like `v1.0.0`, it also publishes those archives as GitHub Release assets.

### Release steps

```bash
git tag v0.1.0
git push origin v0.1.0
```

After the workflow completes, download binaries from the release page.

## Module contract

Every tool module implements `internal/modules.Tool`:

```go
type Tool interface {
    ID() string
    Label() string
    Description() string
    Menu() *menu.Menu
    Requirements() []requirements.Check
}
```

- `Requirements()` runs before entering that module.
- `Menu()` can return deeply nested menus using `menu.NewBuilder(...)`.

## Common menu pattern

Use the builder for consistency:

```go
submenu := menu.NewBuilder("X / Sub").
    Action("Do thing", "Runs action", runFn).
    WithBack().
    Build()

root := menu.NewBuilder("X Tool").
    Action("Top action", "Description", runFn).
    SubMenu("Sub tools", "Nested menu", submenu, optionalChecks...).
    WithBack().
    Build()
```

Use shared exit helpers in any menu:

- `menu.WithBack(items)` or `builder.WithBack()`
- `menu.WithQuit(items)` or `builder.WithQuit()`

## Example modules

- `Hello Tool`
- `Environment Info`
- `Chuck Norris Fact` (calls `https://api.chucknorris.io/jokes/random`)
- `Auth Token Generator` (`Username + Password` and `Google` flows)
- `Cloud CLI Checks` (`AWS CLI` / `Azure CLI` checks with install actions)
# go-devtools
