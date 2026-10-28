# syntax=docker/dockerfile:1

FROM golang:1.27.3 AS builder
WORKDIR /app
# Install build dependencies for Fyne.
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        build-essential \
        pkg-config \
        xorg-dev \
        libwayland-dev \
        libxkbcommon-dev \
        libgl1 \
        libgl1-mesa-dev \
        libegl1 \
        libwayland-client0 \
        libwayland-cursor0 \
        libwayland-egl1 \
        libxkbcommon0 \
        libgl1-mesa-dri && \
    rm -rf /var/lib/apt/lists/*
COPY . .
RUN make build

FROM debian:trixie-slim AS runtime
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        libgl1 \
        libglx0 \
        libgl1-mesa-dri \
        libegl1 \
        libwayland-client0 \
        libwayland-cursor0 \
        libwayland-egl1 \
        libx11-6 \
        libxcursor1 \
        libxi6 \
        libxinerama1 \
        libxrandr2 \
        libxxf86vm1 \
        libxkbcommon0 && \
    rm -rf /var/lib/apt/lists/*
RUN groupadd --system app && \
    useradd --system --gid app --create-home app && \
    mkdir -p /home/app/.cache /run/user && \
    chown -R app:app /home/app
WORKDIR /app
COPY --from=builder --chown=app:app /app/bin/app .
RUN chown -R app:app /app
RUN chmod +x app
USER app
ENTRYPOINT ["./app"]
