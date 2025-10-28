# docker/session/Dockerfile
FROM ubuntu:22.04
ENV DEBIAN_FRONTEND=noninteractive

# ---- Base OS + Neovim -------------------------------------------------------
RUN apt-get update && apt-get install -y --no-install-recommends \
      git curl ca-certificates locales ncurses-term gnupg lsb-release software-properties-common \
  && update-ca-certificates

RUN add-apt-repository -y ppa:neovim-ppa/unstable \
  && apt-get update \
  && apt-get install -y --no-install-recommends neovim \
  && rm -rf /var/lib/apt/lists/*

# Build deps (Treesitter compilers, unzip for some Mason pkgs)
RUN apt-get update && apt-get install -y --no-install-recommends \
      build-essential gcc make pkg-config unzip \
  && rm -rf /var/lib/apt/lists/*

# Locale
RUN locale-gen en_US.UTF-8 && update-locale LANG=en_US.UTF-8
ENV LANG=en_US.UTF-8

# ---- Toolchains --------------------------------------------------------------
# Node + npm (for ts/html/css/json LSPs)
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
  && apt-get install -y --no-install-recommends nodejs \
  && rm -rf /var/lib/apt/lists/*

# Arch-aware Go install (works on amd64 & arm64)
ARG GO_VERSION=1.22.7
ARG TARGETARCH
RUN set -eux; \
  ARCH="${TARGETARCH:-$(dpkg --print-architecture)}"; \
  case "$ARCH" in \
    amd64|x86_64) GOARCH=amd64 ;; \
    arm64|aarch64) GOARCH=arm64 ;; \
    *) echo "Unsupported arch: $ARCH"; exit 1 ;; \
  esac; \
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${GOARCH}.tar.gz" \
  | tar -C /usr/local -xz
ENV PATH="/usr/local/go/bin:${PATH}"
# safer for gopls compile path
ENV CGO_ENABLED=0

# ---- Non-root user & env ----------------------------------------------------
RUN groupadd -r app && useradd -m -r -g app app
ENV HOME=/home/app
ENV TERM=xterm-256color
ENV COLORTERM=truecolor
ENV XDG_CONFIG_HOME=/home/app/.config
ENV XDG_DATA_HOME=/home/app/.local/share
ENV XDG_CACHE_HOME=/home/app/.cache
ENV XDG_STATE_HOME=/home/app/.local/state
RUN mkdir -p /workspace /home/app/.local/share /home/app/.local/state/nvim /home/app/.cache \
  && chown -R app:app /workspace /home/app

# Neovim config (must include mason.nvim + mason-lspconfig; tool-installer optional)
COPY --chown=app:app ./nvim-config/ /home/app/.config/nvim/
RUN nvim --version | head -n1

# ---- Prebake: plugins, LSPs, Treesitter -------------------------------------
USER app
WORKDIR /workspace

# Make runtime skip auto-install; set Go env & PATH for installers
ENV MASON_PREBAKED=1
ENV GOPATH=/home/app/go
ENV PATH="/home/app/.local/share/nvim/mason/bin:${GOPATH}/bin:/usr/local/go/bin:${PATH}"

# 1) Sync plugins so Mason is available
RUN nvim --headless "+Lazy! sync" "+qall"

# 2) Install LSP servers (do both: tool-installer if present, plus explicit list)
RUN nvim --headless "+MasonUpdate" "+MasonToolsInstall" "+qall" || true
RUN nvim --headless "+MasonInstall gopls typescript-language-server html-lsp css-lsp json-lsp lua-language-server" "+qall" || true

# 3) Build Treesitter parsers inside the image
RUN nvim --headless "+TSUpdateSync lua go javascript html css json" "+qall" || true

# 4) Sanity (show what got installed)
RUN ls -l "$HOME/.local/share/nvim/mason/bin" || true
RUN node -v && npm -v && go version || true

# ---- Default command ---------------------------------------------------------
CMD ["nvim"]

