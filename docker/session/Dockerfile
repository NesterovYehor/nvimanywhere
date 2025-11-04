FROM allaman/nvim-full:latest

USER root

ARG GO_VERSION=1.22.7
ARG TARGETARCH

RUN set -eux; \
    arch="${TARGETARCH:-$(dpkg --print-architecture)}"; \
    case "$arch" in \
      amd64|x86_64) goarch=amd64 ;; \
      arm64|aarch64) goarch=arm64 ;; \
      *) echo "Unsupported arch: $arch" >&2; exit 1 ;; \
    esac; \
    apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates curl git \
    && rm -rf /var/lib/apt/lists/*; \
    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${goarch}.tar.gz" \
      | tar -C /usr/local -xz; \
    /usr/local/go/bin/go version

ENV GOPATH=/home/nvim/go
ENV PATH=/usr/local/go/bin:${GOPATH}/bin:${PATH}
RUN mkdir -p /home/nvim/go && chown -R nvim:nvim /home/nvim/go

ENV HOME=/home/nvim
ENV XDG_CONFIG_HOME=/home/nvim/.config
ENV XDG_DATA_HOME=/home/nvim/.local/share
ENV XDG_STATE_HOME=/home/nvim/.local/state
ENV XDG_CACHE_HOME=/home/nvim/.cache

USER nvim

ENV PATH="/home/nvim/.local/share/nvim/mason/bin:${PATH}"

WORKDIR /home/nvim/.config/nvim
COPY --chown=nvim:nvim ./nvim-config/ .

RUN nvim --headless "+Lazy! sync" "+qall"

WORKDIR /workspace

CMD ["nvim"]

