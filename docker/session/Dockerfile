# docker/session/Dockerfile
FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

# Base deps + tools required for add-apt-repository
RUN apt-get update && apt-get install -y --no-install-recommends \
      git curl ca-certificates locales ncurses-term gnupg lsb-release software-properties-common \
  && update-ca-certificates

# Install **latest NVIM (>=0.10)** from the official UNSTABLE PPA (tracks 0.10/0.11)
RUN add-apt-repository -y ppa:neovim-ppa/unstable \
  && apt-get update \
  && apt-get install -y --no-install-recommends neovim \
  && rm -rf /var/lib/apt/lists/*

RUN apt-get update && apt-get install -y \
    build-essential \
    gcc \
    make \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*


# UTF-8 locale
RUN locale-gen en_US.UTF-8 && update-locale LANG=en_US.UTF-8
ENV LANG=en_US.UTF-8

# Create non-root user 'app' with home /home/app
RUN groupadd -r app && useradd -m -r -g app app

# Copy your Neovim config (ensure nvim-config/ is inside the build context)
COPY --chown=app:app ./nvim-config/ /home/app/.config/nvim/

# Terminal + XDG dirs (include STATE so shada/caches can be written)
ENV HOME=/home/app
ENV TERM=xterm-256color
ENV COLORTERM=truecolor
ENV XDG_CONFIG_HOME=/home/app/.config
ENV XDG_DATA_HOME=/home/app/.local/share
ENV XDG_CACHE_HOME=/home/app/.cache
ENV XDG_STATE_HOME=/home/app/.local/state

# Prepare writable dirs and ownership
RUN mkdir -p /workspace \
             /home/app/.local/share \
             /home/app/.local/state/nvim \
             /home/app/.cache \
  && chown -R app:app /workspace /home/app

# Build-time sanity check (prints NVIM version in logs, should be 0.10.x or 0.11-dev)
RUN nvim --version | head -n1

# Run as non-root
USER app
WORKDIR /workspace

# Default command (your Go gateway can still override Cmd when starting the container)
CMD ["nvim"]

