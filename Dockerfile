FROM golang:1.26-bookworm AS builder

# Dependencias de CGO para Ebiten (X11, ALSA, OpenGL)
RUN apt-get update && apt-get install -y \
    libasound2-dev \
    libx11-dev \
    libxrandr-dev \
    libxinerama-dev \
    libxcursor-dev \
    libxi-dev \
    libgl1-mesa-dev \
    libxxf86vm-dev \
    pkg-config \
    gcc \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /usr/local/bin/bgp-viewer ./cmd/bgp-viewer
RUN go build -o /usr/local/bin/bgp-cli ./cmd/bgp-cli

# Imagen final
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    libasound2 \
    libx11-6 \
    libxrandr2 \
    libxinerama1 \
    libxcursor1 \
    libxi6 \
    libgl1 \
    libxxf86vm1 \
    xvfb \
    ffmpeg \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /usr/local/bin/bgp-viewer /usr/local/bin/bgp-viewer
COPY --from=builder /usr/local/bin/bgp-cli /usr/local/bin/bgp-cli

WORKDIR /data
VOLUME ["/data"]

CMD ["/usr/local/bin/bgp-viewer"]