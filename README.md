# better-idn

This repository is the implementation of the better-idn backend. Built with:

- Go 1.24.0
- PostgreSQL
- Golang Migrate
- Air for live reloading

## Installation

### Prerequisites
Make sure you have the following installed:
- [Go](https://golang.org/dl/)
- [PostgreSQL](https://www.postgresql.org/download/)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [Air](https://github.com/air-verse/air) for live reloading

### Install Go and Required Packages

If Go is installed via Homebrew, it does not automatically set up `GOPATH`. Configure it manually:

```sh
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

Then, install `golang-migrate` and `Air`:

```sh
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/air-verse/air@latest
```

## Project Setup

Clone the repository:

```sh
git clone https://github.com/thediligencedev/better-idn.git
cd better-idn
```

Set up environment variables by copying `.env.example` to `.env` and updating it with your PostgreSQL credentials:

```sh
cp .env.example .env
nano .env
```

## Running the Project

### Using Air for Live Reloading

Ensure `air.toml` is set up correctly. The configuration should look like this:

```toml
root = "."

[build]
  cmd = "go build -o ./tmp/main ./cmd/server/main.go"
  bin = "tmp/main"
  full_bin = "tmp/main"
  include_ext = ["go", "tpl", "html"]
  exclude_dir = ["assets", "tmp"]
  delay = 1000

[run]
  cmd = "./tmp/main"
  dir = "."
```

Run the server with live reload:

```sh
air
```

### Documentation Page

You can access the API documentation page that served inside swagger in `http://localhost:8080/api/docs`

### Demo Using HTMX
You can serve the HTMX frontend with `cd frontend-demo` then `python3 -m http.server 5500`

### Using Makefile

The project includes a `Makefile` for easy build and database migration management:

#### Build the project:
```sh
make build
```

#### Run the project (without live reloading, better use air):
```sh
make run
```

#### Run database migrations:

Apply migrations:
```sh
make migrate-up
```

Rollback migrations:
```sh
make migrate-down
```

Create a new migration file:
```sh
make migrate-create
```
You will be prompted to enter a migration name.

## Testing
Run tests with:
```sh
make test
```

## Code Quality
Format and lint the code:
```sh
make fmt
make vet
```

## Cleanup
Remove build artifacts:
```sh
make clean
```

---

This repository is structured to allow seamless development and deployment, integrating best practices for Go projects.
