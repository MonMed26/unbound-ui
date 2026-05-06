# Unbound UI

Web-based management interface for [Unbound](https://nlnetlabs.nl/projects/unbound/about/) DNS server.

![License](https://img.shields.io/badge/license-MIT-blue.svg)

## Features

- **Dashboard** - Real-time statistics (queries, cache hit ratio, uptime, memory)
- **Configuration Editor** - Edit unbound.conf with validation and backup
- **Zone Management** - Manage local zones and DNS records via UI
- **DNS Blocklist** - Block ads/trackers/malware with multiple source support
- **Authentication** - Secure access with username/password + JWT sessions

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Frontend | React 18 + Vite + TypeScript |
| UI | Tailwind CSS + Custom Components |
| Backend | Go 1.22 + Chi Router |
| Auth | bcrypt + JWT |

## Quick Start

### Docker (Recommended)

```bash
cd docker
docker compose up -d
```

Access the UI at `http://localhost:8080`

### Bare Metal

**Prerequisites:**
- Go 1.22+
- Node.js 20+
- Unbound DNS server installed

**Build:**

```bash
make build
```

**Install:**

```bash
sudo ./scripts/install.sh
```

## Development

**Frontend (port 3000):**

```bash
cd frontend
npm install
npm run dev
```

**Backend (port 8080):**

```bash
cd backend
go run ./cmd/server
```

The frontend dev server proxies `/api` requests to the backend.

## Configuration

Configuration file: `/etc/unbound-ui/config.yaml` (or `config.yaml` in working directory)

```yaml
server:
  port: 8080
  host: "0.0.0.0"

unbound:
  config_path: "/etc/unbound/unbound.conf"
  control_path: "unbound-control"

auth:
  username: "admin"
  password_hash: "$2a$10$..."
  jwt_secret: "auto-generated"
  session_ttl: "24h"

blocklist:
  data_dir: "/var/lib/unbound-ui/blocklist"
  output_path: "/etc/unbound/unbound.conf.d/blocklist.conf"
  update_interval: "6h"
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `UNBOUND_UI_PORT` | Server port | 8080 |
| `UNBOUND_UI_HOST` | Bind address | 0.0.0.0 |
| `UNBOUND_CONFIG_PATH` | Path to unbound.conf | /etc/unbound/unbound.conf |
| `UNBOUND_CONTROL_PATH` | Path to unbound-control | unbound-control |

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/login` | Login |
| POST | `/api/auth/setup` | First-time setup |
| GET | `/api/stats` | Dashboard statistics |
| GET | `/api/config` | Get configuration |
| PUT | `/api/config` | Update configuration |
| POST | `/api/config/reload` | Reload unbound |
| GET | `/api/zones` | List local zones |
| POST | `/api/zones` | Add zone |
| DELETE | `/api/zones/:name` | Remove zone |
| GET | `/api/blocklist/sources` | List blocklist sources |
| POST | `/api/blocklist/sources` | Add source |
| POST | `/api/blocklist/update` | Update all blocklists |
| POST | `/api/cache/flush` | Flush DNS cache |

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   React SPA     │────▶│   Go API Server  │────▶│  Unbound DNS    │
│   (Browser)     │     │   (port 8080)    │     │  (port 53)      │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                              │                        │
                              ├── unbound-control ─────┘
                              ├── read/write unbound.conf
                              └── manage blocklists
```

## License

MIT
