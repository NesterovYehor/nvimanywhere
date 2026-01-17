#  NvimAnywhere

**NvimAnywhere** is a lightweight, self-hostable service that lets you run **your Neovim environment directly in the browser**, backed by isolated Docker containers.

Each browser session spawns a dedicated container running Neovim, with terminal I/O streamed over WebSockets and rendered using **xterm.js**. Sessions are ephemeral, isolated, and cleaned up automatically.

> 💡 The goal is not to replace SSH — but to provide a predictable, reproducible Neovim workspace accessible from any device with a browser.

---

## ✨ Features

* One **Docker container per session**
* Real Neovim TUI streamed to the browser
* Uses your **existing Neovim configuration**
* Deterministic session lifecycle and cleanup
* No persistent state unless explicitly mounted
* Fully self-hostable

---

## 🧠 Motivation

Local development environments are often:

* inconsistent across machines
* hard to reproduce
* unsafe or inconvenient in shared or restricted environments

NvimAnywhere explores an alternative model:

* ephemeral development sessions
* strict container isolation
* explicit lifecycle ownership

This makes it suitable for:

* working from shared or locked-down machines
* demos or workshops
* experimenting with remote development setups
* learning systems design around PTYs, WebSockets, and containers

---

## ⚙️ How It Works

1. The user opens a new session in the browser.
2. The backend generates a **unique session token**.
3. A Docker container is started for that session.
4. Neovim runs inside the container, attached to a PTY.
5. A WebSocket bridge streams terminal I/O between the browser and the container.
6. **xterm.js** renders the Neovim TUI in the browser.
7. On disconnect, the container and workspace are cleaned up automatically.

---

## 🧩 Architecture

```text
Browser (xterm.js)
        │
        │  WebSocket (binary)
        ▼
Go Gateway (HTTP + WS)
        │
        │  Docker API
        ▼
Session Container
(Neovim + PTY)
```

---

## 🛠️ Tech Stack

* **Backend:** Go (net/http, gorilla webSockets, Docker SDK)
* **Frontend:** xterm.js + HTML/CSS/JS
* **Isolation:** Docker (one container per session)
* **Logging:** slog (structured logging)
* **Configuration:** YAML with environment overrides
* **Assets:** embed.FS

---

## 🚀 Getting Started

### Prerequisites

* Docker ≥ 24
* Linux or macOS
* Go ≥ 1.22 (optional, for local builds)

---

### Configuration

Create a configuration file (for example `config.yaml`):

```yaml
http:
  host: "0.0.0.0"
  port: "8080"

session_runtime:
  image_name: "nvimanywhere-session:latest"
  base_path: "/srv/nvimanywhere/data/workspaces"

log_file_path: ""
env: "production"
```

---

### Running with Docker

```bash
docker build -t nvimanywhere:latest .

mkdir -p /srv/nvimanywhere/data/workspaces

docker run -d \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /srv/nvimanywhere/data/workspaces:/srv/nvimanywhere/data/workspaces \
  -v $(pwd)/config.yaml:/etc/nva/config.yaml:ro \
  -e NVA_CONFIG=/etc/nva/config.yaml \
  nvimanywhere:latest
```

Then open:

```
http://localhost:8080
```

---

### Running Locally (without Docker image build)

```bash
go build -o nvimanywhere ./cmd/gateway
export NVA_CONFIG=$(pwd)/config.yaml
./nvimanywhere
```

---

## ⚠️ Notes & Limitations

* Clipboard integration relies on browser selection; native clipboard sync is not implemented in V1.
* Each session is isolated and ephemeral by design.
* Intended as a developer tool / learning project, not a hosted SaaS.

---

## 📜 License

MIT License © 2025 Yehor Nesterov
