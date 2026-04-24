# Make Commands

Quick helper commands for local development.

## Prerequisite

- Docker + Docker Compose v2 installed
- Run commands from the repository root

## Commands

| Command | What it does |
|---|---|
| `make dev_up` | Build and start the full stack in background |
| `make dev_down` | Stop and remove containers (keeps volumes) |
| `make dev_rebuild` | Recreate everything (down + build + up) |
| `make dev_logs` | Follow logs from all services |
| `make dev_ps` | Show current service status |
| `make dev_seed` | Seed demo data |
| `make dev_reset` | Reset non-demo data |

## Typical flow

```bash
make dev_up
make dev_seed
make dev_logs
```

## Notes

- Services started by `dev_up`: postgres, minio, app, frontend.
- `dev_seed` and `dev_reset` are one-off utility jobs.
