# gmx

**Go Main Expansion**: a command-line scaffolder that generates a development-ready [Gin](https://gin-gonic.com/) backend in one command, with optional [Supabase](https://supabase.com/) integration.

```bash
gmx build myapi --db supabase
```

## Features

- Clean, layered project structure (`http` → `routes` → `service` → `database`)
- CORS configured through environment variables
- Optional Supabase database layer behind an `IDatabase` interface, so it is easy to mock and swap
- Health-check endpoints (`/test/ping`, `/test/db`) to verify your setup
- Tests included for the server, routes and services
- `.env.example` and `.gitignore` generated for you

## Requirements

- [Go](https://go.dev/dl/) 1.22 or newer
- A [Supabase](https://supabase.com/) project (only when using `--db supabase`)

## Installation

```bash
go install github.com/IsaacAgyei/gmx@latest
```

Make sure `$(go env GOPATH)/bin` is on your `PATH`, then verify the install:

```bash
gmx version
```

## Quick start with Supabase

### 1. Generate the project

```bash
gmx build myapi --db supabase
cd myapi
```

This creates the project, runs `go mod init`, installs dependencies and formats the code.

### 2. Get your Supabase credentials

In the [Supabase dashboard](https://supabase.com/dashboard), open your project and go to **Project Settings → API**. Copy:

| Value | Used for |
| --- | --- |
| **Project URL** | `SUPABASE_URL` |
| **API key** (`anon` for public access, `service_role` for trusted servers only) | `SUPABASE_KEY` |

> **Security:** the `service_role` key bypasses Row Level Security. Use it only on a trusted server and never expose it to a browser or commit it to git.

### 3. Configure the environment

```bash
cp .env.example .env
```

Edit `.env`:

```env
ALLOWED_ORIGINS=http://localhost:4200
SUPABASE_URL=https://your-project-ref.supabase.co
SUPABASE_KEY=your-api-key
```

`.env` is git-ignored. Never commit it.

### 4. Create the `user_profile` table

The generated database layer reads from a `user_profile` table, and `/test/db` uses it to check connectivity. Run this in the Supabase **SQL Editor**:

```sql
create table if not exists user_profile (
  user_id  uuid primary key default gen_random_uuid(),
  username text unique not null
);
```

If Row Level Security is enabled on the table, add a policy that allows the key you chose to read it, or use the `service_role` key.

### 5. Run the server

```bash
go run .
```

### 6. Verify

```bash
curl localhost:8080/api/v1/test/ping   # server is up
curl localhost:8080/api/v1/test/db     # Supabase is reachable
```

## Command reference

```text
gmx build [dir] [flags]    create a new Gin project in dir (default: backend)
gmx version                print the gmx version
gmx help                   show help
```

| Flag | Default | Description |
| --- | --- | --- |
| `--db <provider>` | none | Add a database layer. Supported: `supabase` |
| `--module <path>` | directory name | Go module path of the generated project |
| `--prefix <path>` | `/api/v1` | Parent route group prefix |
| `--skip-install` | off | Skip `go get` / `go mod tidy` (no network needed) |
| `--force` | off | Write into a non-empty directory |

### Examples

```bash
gmx build                                   # ./backend (module "backend")
gmx build myapi                             # ./myapi   (module "myapi")
gmx build myapi --db supabase               # with Supabase
gmx build api --module github.com/me/api --prefix /api/v2
gmx build myapi --skip-install              # offline
```

## What gets generated

With `--db supabase`:

```text
myapi/
├── main.go                  loads .env, connects to the database, starts the server
├── .env.example             ALLOWED_ORIGINS, SUPABASE_URL, SUPABASE_KEY
├── database/
│   ├── database.go          IDatabase interface and shared types
│   └── supabase/            Supabase implementation
├── http/                    Gin engine and CORS
├── routes/                  route groups and the ping/db test routes
└── service/                 business logic used by the routes
```

Without `--db`, the `database/` directory and the `/test/db` route are omitted. The generated project also contains its own README with details on adding routes.

## Troubleshooting

| Problem | Fix |
| --- | --- |
| `gmx: command not found` | Add `$(go env GOPATH)/bin` to your `PATH`. |
| `<dir> is not empty` | Choose a new directory or pass `--force`. |
| `SUPABASE_URL and SUPABASE_KEY must be set` | Create `.env` from `.env.example` and fill in both values. |
| `/test/db` returns an error | Check the URL and key, and that the `user_profile` table exists and is readable under your RLS policies. |
| Dependency install failed | Re-run `go get ./... && go mod tidy` inside the project when online. |

## Customizing (for forks)

Everything generated lives in `templates/` as `*.go.tmpl` files (`text/template`, with `{{.Module}}`, `{{.Prefix}}` and `{{.DB}}`). Edit or add a file there and rebuild gmx; templates are embedded with `go:embed`. A leading `dot_` in a file name becomes `.` (for example `dot_gitignore.tmpl` → `.gitignore`). Files under `templates/database/<provider>/` are only generated when that provider is selected with `--db`.

## Contributions

This project is open source but is not accepting pull requests or issues. You are free to fork it under the license terms.

## License

Copyright (C) 2026 Isaac Agyei

gmx is licensed under the GNU General Public License v3.0. See [LICENSE](LICENSE).

Original source: <https://github.com/IsaacAgyei/gmx>

Modified versions must carry prominent notices stating they were changed, and must be released under the same license.

**Generated output:** as an additional permission under GPLv3 section 7, the files that `gmx build` generates are not covered by the GPL. You may use, modify and license them however you like, including in proprietary projects.
