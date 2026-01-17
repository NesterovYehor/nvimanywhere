FROM allaman/nvim-full:latest

USER root

ARG GO_VERSION=1.22.7
ARG TARGETARCH

RUN set -eux; \
    arch="${TARGETARCH:-$(dpkg --print-architecture)}"; \
    case "$arch" in \
      amd64|x86_64) goarch=amd64 ;; \
      arm64|aarch64) goarch=arm64 ;; \
      *) exit 1 ;; \
    esac; \
    apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl git \
    && rm -rf /var/lib/apt/lists/*; \
    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${goarch}.tar.gz" \
      | tar -C /usr/local -xz

ENV HOME=/home/nvim
ENV GOPATH=/home/nvim/go
ENV PATH=/usr/local/go/bin:${GOPATH}/bin:${PATH}

ENV XDG_CONFIG_HOME=/home/nvim/.config
ENV XDG_DATA_HOME=/home/nvim/.local/share
ENV XDG_STATE_HOME=/home/nvim/.local/state
ENV XDG_CACHE_HOME=/home/nvim/.cache

RUN mkdir -p /home/nvim/go && chown -R nvim:nvim /home/nvim

# --- bootstrap plugins as root ---
COPY docker/session/base-nvim-config/ /tmp/base-nvim-config

RUN XDG_CONFIG_HOME=/tmp \
    NVIM_APPNAME=nvim \
    nvim --headless \
      "+Lazy! sync" \
      "+MasonInstallAll" \
      "+qall"

RUN rm -rf /tmp/base-nvim-config

RUN mkdir -p /home/nvim/.cache /home/nvim/.local \
 && chown -R nvim:nvim /home/nvim/.cache /home/nvim/.local

# --- runtime ---
USER nvim
ENV PATH="/home/nvim/.local/share/nvim/mason/bin:${PATH}"

WORKDIR /workspace
CMD ["nvim", "."]

