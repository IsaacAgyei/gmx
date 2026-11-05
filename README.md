# gmx > go-main-expansion

Scaffolds a Go + Gin backend: `gmx build`.

## Install

Requires Go 1.22 or newer.

```bash
go install github.com/IsaacAgyei/gmx@latest
```

Make sure `$(go env GOPATH)/bin` is on your PATH, then run `gmx version` to check.

## Use

```bash
gmx build                      # creates ./backend  (module "backend")
gmx build myapi                # creates ./myapi    (module "myapi")
gmx build api --module github.com/me/api --prefix /api/v2
gmx build --skip-install       # don't run go get / go mod tidy
gmx build myapi --db supabase  # also adds a Supabase database layer (see below)
```

### Database (`--db supabase`)

Adds `database/` (an `IDatabase` interface plus a Supabase implementation), wires it through
`main.go` -> server -> routes -> services, adds `GET <prefix>/test/db`, and puts `SUPABASE_URL` and
`SUPABASE_KEY` in `.env.example`. Copy it to `.env` and fill in your values; never commit `.env`.

`gmx build` writes the project, runs `go mod init`, then `go get` for gin, gin-contrib/cors and godotenv,
then `go mod tidy`. See the generated README for the layout.

## Customizing (for forks)

Everything generated lives in `templates/` as `*.go.tmpl` files (`text/template`, with `{{.Module}}`
and `{{.Prefix}}`). Edit or add a file there and rebuild gmx; it is embedded with `go:embed`.
A leading `dot_` in a file name becomes `.` (e.g. `dot_gitignore.tmpl` -> `.gitignore`).

## Contributions

This project is open source but is not accepting pull requests or issues.
You're free to fork it under the license terms.

## License

Copyright (C) 2026 Isaac Agyei

gmx is licensed under the GNU General Public License v3.0. See [LICENSE](LICENSE).

Original source: [https://github.com/IsaacAgyei/gmx](https://github.com/IsaacAgyei/gmx)

Modified versions must carry prominent notices stating they were changed, and must be released under the same license.

**Generated output:** As an additional permission under GPLv3 section 7, the files
that `gmx build` generates are not covered by the GPL. You may use, modify and
license them however you like, including in proprietary projects.
