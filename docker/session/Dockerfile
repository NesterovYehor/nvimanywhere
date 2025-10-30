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
    apt-get update && apt-get install -y --no-install-recommends ca-certificates curl && rm -rf /var/lib/apt/lists/*; \
    curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-${goarch}.tar.gz" \
      | tar -C /usr/local -xz; \
    /usr/local/go/bin/go version

# Set Go env for the nvim user
ENV GOPATH=/home/nvim/go
ENV PATH=/usr/local/go/bin:${GOPATH}/bin:${PATH}
RUN mkdir -p /home/nvim/go && chown -R nvim:nvim /home/nvim/go
# ------------------------------------------------------------------------------

# switch back to the non-root user your image uses
USER nvim

# run as the non-root user the image provides
USER nvim

# (optional, but helpful) ensure mason’s bin is on PATH
ENV PATH="/home/nvim/.local/share/nvim/mason/bin:${PATH}"


# copy ONLY your config folder into the correct place
WORKDIR /home/nvim/.config/nvim
COPY --chown=nvim:nvim docker/session/nvim-config/ .

# install plugins
RUN nvim --headless "+Lazy! sync" "+qall"

# where you want to edit files
WORKDIR /workspace

CMD ["nvim"]

