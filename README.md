# gmx

**Go Main Expansion**: a command-line scaffolder that generates a development-ready [Gin](https://gin-gonic.com/) backend in one command, with optional [Supabase](https://supabase.com/) and [Stripe](https://stripe.com/) integration.

```bash
gmx build myapi --db supabase --payment stripe
```

## Features

- Clean, layered project structure (`http` → `routes` → `service` → `database` / `paymentprocessor`)
- CORS configured through environment variables
- Optional Supabase database layer behind an `IDatabase` interface, so it is easy to mock and swap
- Optional Stripe payment layer behind an `IPaymentProcessor` interface, with a `POST /payments/payment` route that takes a JSON body
- Health-check endpoints (`/test/ping`, `/test/db`) to verify your setup
- Tests included for the server, routes and services
- `.env.example` and `.gitignore` generated for you

## Requirements

- [Go](https://go.dev/dl/) 1.22 or newer
- A [Supabase](https://supabase.com/) project (only when using `--db supabase`)
- A [Stripe](https://stripe.com/) account and a test secret key (only when using `--payment stripe`)

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

| Value                                                                           | Used for       |
| ------------------------------------------------------------------------------- | -------------- |
| **Project URL**                                                                 | `SUPABASE_URL` |
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
STRIPE_TEST_KEY=sk_test_...   # only when generated with --payment stripe
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

### 5. Build and run the server

```bash
go build .
go run .
```

### 6. Verify

```bash
curl localhost:8080/api/v1/test/ping   # server is up
curl localhost:8080/api/v1/test/db     # Supabase is reachable
```

## Adding Stripe payments

```bash
gmx build myapi --payment stripe          # Stripe only
gmx build myapi --db supabase --payment stripe # Supabase and Stripe
```

Get a secret key from the [Stripe dashboard](https://dashboard.stripe.com/test/apikeys) (use a `sk_test_...` key while developing) and set it in `.env`:

```env
STRIPE_TEST_KEY=sk_test_...
```

> **Security:** never commit your Stripe secret key or expose it to a browser. `.env` is git-ignored.

The server refuses to start if `STRIPE_TEST_KEY` is empty. Then create a payment with a JSON body whose field names match the Stripe PaymentIntent parameters:

```bash
curl -X POST localhost:8080/api/v1/payments/payment \
  -H "Content-Type: application/json" \
  -d '{"amount": 2000, "currency": "usd", "payment_method": "pm_card_visa"}'
```

| Field            | Required | Description                                                                                                        |
| ---------------- | -------- | ------------------------------------------------------------------------------------------------------------------ |
| `amount`         | yes      | Smallest currency unit, greater than 0 (`2000` = $20.00)                                                           |
| `currency`       | yes      | Three-letter ISO code, e.g. `usd`                                                                                  |
| `payment_method` | yes      | `pm_card_visa` for testing, or a payment method id from Stripe.js                                                  |
| `return_url`     | no       | Valid URL; allows redirect-based payment methods. When omitted, only non-redirect methods (e.g. cards) are allowed |

An invalid body returns `400`; a failed or declined payment returns `502`.

## Command reference

```text
gmx build [dir] [flags]    create a new Gin project in dir (default: backend)
gmx version                print the gmx version
gmx help                   show help
```

| Flag                    | Default        | Description                                        |
| ----------------------- | -------------- | -------------------------------------------------- |
| `--db <provider>`       | none           | Add a database layer. Supported: `supabase`        |
| `--payment <processor>` | none           | Add a payment processor layer. Supported: `stripe` |
| `--module <path>`       | directory name | Go module path of the generated project            |
| `--prefix <path>`       | `/api/v1`      | Parent route group prefix                          |
| `--skip-install`        | off            | Skip `go get` / `go mod tidy` (no network needed)  |
| `--force`               | off            | Write into a non-empty directory                   |

### Examples

```bash
gmx build                                   # ./backend (module "backend")
gmx build myapi                             # ./myapi   (module "myapi")
gmx build myapi --db supabase               # with Supabase
gmx build myapi --payment stripe            # with Stripe
gmx build myapi --db supabase --payment stripe
gmx build api --module github.com/me/api --prefix /api/v2
gmx build myapi --skip-install              # offline
```

## What gets generated

With `--db supabase --payment stripe`:

```text
myapi/
├── main.go                  loads .env, connects the providers, starts the server
├── .env.example             ALLOWED_ORIGINS, SUPABASE_URL, SUPABASE_KEY, STRIPE_TEST_KEY
├── database/
│   ├── database.client.interface.go   IDatabase interface and shared types
│   └── supabase/
│       └── supabase.go                Supabase implementation
├── paymentprocessor/
│   ├── interface.payment.processor.go IPaymentProcessor interface and PaymentRequest
│   └── stripe/
│       └── stripe.go                  Stripe implementation
├── http/                    Gin engine and CORS
├── routes/
│   ├── routes.register.go   parent route group, registers every route
│   ├── ping/                GET /test/ping and /test/db
│   └── payments/            POST /payments/payment
└── service/
    ├── service.config.go              Deps, Services registry and constructor
    ├── service.server.ping.go         ServerPingService
    ├── service.database.ping.go       DatabasePingService
    └── service.paymentprocessor.go    PaymentProcessorService
```

Without `--db`, the `database/` directory, `service.database.ping.go` and the `/test/db` route are omitted. Without `--payment`, the `paymentprocessor/` directory, `routes/payments/` and `service.paymentprocessor.go` are omitted. The generated project also contains its own README with details on adding routes.

## Troubleshooting

| Problem                                     | Fix                                                                                                                  |
| ------------------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `gmx: command not found`                    | Add `$(go env GOPATH)/bin` to your `PATH`.                                                                           |
| `<dir> is not empty`                        | Choose a new directory or pass `--force`.                                                                            |
| `SUPABASE_URL and SUPABASE_KEY must be set` | Create `.env` from `.env.example` and fill in both values.                                                           |
| `STRIPE_TEST_KEY must be set`               | Create `.env` from `.env.example` and set your Stripe secret key.                                                    |
| `/payments/payment` returns `400`           | The body must be JSON with `amount`, `currency` and `payment_method`, and a `Content-Type: application/json` header. |
| `/payments/payment` returns `502`           | The payment failed or Stripe rejected the request; check the key and the server log for the Stripe error.            |
| `/test/db` returns an error                 | Check the URL and key, and that the `user_profile` table exists and is readable under your RLS policies.             |
| Dependency install failed                   | Re-run `go get ./... && go mod tidy` inside the project when online.                                                 |

## Customizing (for forks)

Everything generated lives in `templates/` as `*.go.tmpl` files (`text/template`, with `{{.Module}}`, `{{.Prefix}}`, `{{.DB}}` and `{{.Payment}}`). Edit or add a file there and rebuild gmx; templates are embedded with `go:embed`. A leading `dot_` in a file name becomes `.` (for example `dot_gitignore.tmpl` → `.gitignore`). Files under `templates/database/<provider>/` are only generated when that provider is selected with `--db`. The same applies to `templates/paymentprocessor/<provider>/` and `--payment`. Which files belong to which flag is decided by `enabled()` in [build.go](build.go).

## Contributions

This project is open source but is not accepting pull requests or issues. You are free to fork it under the license terms.

## License

Copyright (C) 2026 Isaac Agyei

gmx is licensed under the GNU General Public License v3.0. See [LICENSE](LICENSE).

Original source: <https://github.com/IsaacAgyei/gmx>

Modified versions must carry prominent notices stating they were changed, and must be released under the same license.

**Generated output:** as an additional permission under GPLv3 section 7, the files that `gmx build` generates are not covered by the GPL. You may use, modify and license them however you like, including in proprietary projects.
