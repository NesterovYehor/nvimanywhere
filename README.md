# 🌀 NvimAnywhere

Traditional local development environments are inconsistent across machines, require setup, and complicate resource management in multi-tenant environments. NvimAnywhere addresses this by providing containerized, ephemeral dev workspaces with predictable behavior.

**NvimAnywhere** is a lightweight, self-hostable service that lets you open and use your personal **Neovim setup directly from the browser** — from anywhere, on any device.  
It provides a secure, container-isolated environment running Neovim inside Docker, streamed through a WebSocket bridge with a real terminal interface built on **xterm.js**.

> ⚡ Use your Neovim config, plugins, and workflows outside your main machine — e.g., on campus or any remote device with internet access.

---

## 📑 Table of Contents

- [Overview](#-overview)
- [How It Works](#-how-it-works)
- [Architecture Diagram](#-architecture-diagram)
- [Tech Stack](#-tech-stack)
- [Installation](#-installation)
  - [1. Prerequisites](#1️⃣-prerequisites)
  - [2. Configuration](#2️⃣-configuration)
  - [3. Build and Run](#3️⃣-build-and-run)
- [Configuration Reference](#-configuration-reference)
- [License](#-license)

---

## 🌍 Overview

NvimAnywhere replicates your **local Neovim experience** directly in the browser — complete with syntax highlighting, terminal capabilities, and isolated workspaces — while ensuring reproducibility and resource safety.

Each browser session:
- Spawns a **dedicated Docker container** running Neovim.
- Creates a **WebSocket bridge** between the browser and container PTY.
- Streams terminal I/O in real time via **xterm.js**.
- Cleans up automatically on disconnect or timeout.

Perfect for:
- Students or developers working across multiple devices.  
- Using Neovim in restricted environments (university labs, shared machines).  
- Demoing or showcasing your custom Neovim setup.

---

## ⚙️ How It Works

1. User opens `/sessions/new` → backend creates a **unique session token**.  
2. A **Docker container** starts, mounted under `/srv/nvimanywhere/data/workspaces/<token>`.  
3. A **WebSocket connection** bridges browser ↔ backend ↔ PTY.  
4. **xterm.js** renders the Neovim TUI in the browser.  
5. On disconnect, everything is cleaned up automatically.

---

## 🧩 Architecture Diagram

```text
                                 ┌──────────────────────────┐
                                 │        Browser           │
                                 │      (xterm.js UI)       │
                                 └────────────┬─────────────┘
                                              │  HTTPS + WebSocket (upgrade)
                                              │
                                 ┌────────────▼─────────────┐
                                 │       Go Gateway         │
                                 │    (HTTP + WS server)    │
                                 │                          │
                                 │  ┌────────────────────┐  │
                                 │  │  Bridge (WS ↔ PTY) │  │
                                 │  │  • single writer   │  │
                                 │  │  • size/time flush │  │
                                 │  │  • write deadlines │  │
                                 │  └─────────┬──────────┘  │
                                 └────────────│─────────────┘
                                              │  local Docker API
                                 ┌────────────▼─────────────┐
                                 │   Session Container      │
                                 │     (Neovim + PTY)       │
                                 │  - one per session       │
                                 │  - resource caps         │
                                 │  - /workspaces bind      │
                                 └──────────────────────────┘
```

---

## 🛠️ Tech Stack

**Backend:** Go (HTTP, WebSocket, Docker SDK)  
**Frontend:** xterm.js + minimal HTML/CSS  
**Isolation:** Docker (one container per session)  
**Logging:** `slog` structured logs  
**Config:** YAML (with optional env overrides)  
**Templates:** `embed.FS`  
**Transport:** Binary WebSocket with buffered coalescing writes  

---

## 🚀 Installation

### 1️⃣ Prerequisites
- **Docker** ≥ 24.x  
- **Go** ≥ 1.22 (optional, for local build)  
- Linux or macOS (tested on both)

---

### 2️⃣ Configuration

Create `config/config.yaml`:

```yaml
http:
  host: "0.0.0.0"
  port: "8080"

container:
  image: "nvimanywhere-session:latest"
  workspaces_path: "/srv/nvimanywhere/data/workspaces"

log_file_path: "./logs/nvimanywhere.log"
env: "production"
```

---

### 3️⃣ Build and Run

#### **Docker**

```bash
sudo mkdir -p /srv/nva/workspaces

docker build -t nvimanywhere:latest .

docker run -d --name nvimanywhere \
  -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /srv/nva/workspaces:/workspaces \
  -v $(pwd)/config/config.yaml:/etc/nva/config.yaml:ro \
  nvimanywhere:latest
```

Then open:

```
http://localhost:8080/
```

#### **Local Binary (optional)**

```bash
go build -o bin/nvimanywhere ./cmd/gateway
export NVA_CONFIG=$(pwd)/config/config.yaml
./bin/nvimanywhere
```

---

## ⚙️ Configuration Reference

| Key | Description | Default | Required |
|-----|--------------|----------|-----------|
| `http.host` | Address to bind HTTP server | `"0.0.0.0"` | No |
| `http.port` | Port to listen on | `"8080"` | No |
| `container.image` | Docker image used for Neovim sessions | — | **Yes** |
| `container.workspaces_path` | Workspace directory inside session container | `"/workspace"` | No |
| `log_file_path` | File path for logs (`stdout` if empty) | `""` | No |
| `env` | Application environment (`dev` / `prod`) | `"prod"` | No |

---

## 📜 License

MIT License © 2025 Yehor Nesterov  
[github.com/NesterovYehor](https://github.com/NesterovYehor)
