#!/bin/sh
set -eu

version=${KANSHI_VERSION:-v1.0.0}
prefix=${PREFIX:-/usr/local}
service=false
# --service installs a boot-time service using the platform's init system
# (systemd on Linux, launchd on macOS). --systemd is kept as an alias.
case "${1:-}" in
  --service | --systemd) service=true ;;
esac

case $(uname -s) in
  Linux) os=linux ;;
  Darwin) os=darwin ;;
  *) echo "unsupported operating system: $(uname -s)" >&2; exit 1 ;;
esac
case $(uname -m) in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

release_version=${version#v}
archive="kanshi-agent_${release_version}_${os}_${arch}.tar.gz"
base="https://github.com/kanshi-dev/agent/releases/download/${version}"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT HUP INT TERM

curl -fsSL "$base/$archive" -o "$tmp/$archive"
curl -fsSL "$base/checksums.txt" -o "$tmp/checksums.txt"
expected=$(awk -v file="$archive" '$2 == file { print $1 }' "$tmp/checksums.txt")
[ -n "$expected" ] || { echo "checksum not found for $archive" >&2; exit 1; }
if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$tmp/$archive" | awk '{print $1}')
else
  actual=$(shasum -a 256 "$tmp/$archive" | awk '{print $1}')
fi
[ "$actual" = "$expected" ] || { echo "checksum verification failed" >&2; exit 1; }

tar -xzf "$tmp/$archive" -C "$tmp"

sudo=""
need_root() {
  [ "$(id -u)" -eq 0 ] && return 0
  if command -v sudo >/dev/null 2>&1; then
    sudo="sudo"
    return 0
  fi
  echo "need root to write to $prefix; re-run as root, or set PREFIX to a writable directory" >&2
  exit 1
}

# Install the binary, escalating with sudo only when the prefix is not writable.
# mkdir -p never re-chmods an existing directory, so a writable-but-unowned
# /usr/local/bin (Homebrew, common macOS setups) installs without a sudo prompt.
mkdir -p "$prefix/bin" 2>/dev/null || true
if [ -w "$prefix/bin" ]; then
  install -m 0755 "$tmp/kanshi-agent" "$prefix/bin/kanshi-agent"
else
  need_root
  $sudo install -d "$prefix/bin"
  $sudo install -m 0755 "$tmp/kanshi-agent" "$prefix/bin/kanshi-agent"
fi

if "$service"; then
  case "$os" in
    linux | darwin) ;;
    *) echo "--service is not supported on $os" >&2; exit 1 ;;
  esac

  # Ask for required config when it is not supplied via the environment.
  if [ -z "${KANSHI_CORE_ADDR:-}" ]; then
    [ -r /dev/tty ] || { echo "set KANSHI_CORE_ADDR (no terminal to prompt)" >&2; exit 1; }
    printf 'Core address (host:50051): ' > /dev/tty
    read -r KANSHI_CORE_ADDR < /dev/tty
  fi
  if [ -z "${KANSHI_API_KEY:-}" ]; then
    [ -r /dev/tty ] || { echo "set KANSHI_API_KEY (no terminal to prompt)" >&2; exit 1; }
    printf 'Ingest API key (KANSHI_API_KEY from your core .env): ' > /dev/tty
    stty -echo < /dev/tty 2>/dev/null || true
    read -r KANSHI_API_KEY < /dev/tty
    stty echo < /dev/tty 2>/dev/null || true
    printf '\n' > /dev/tty
  fi
  [ -n "$KANSHI_CORE_ADDR" ] || { echo "core address is required" >&2; exit 1; }
  [ -n "$KANSHI_API_KEY" ] || { echo "API key is required" >&2; exit 1; }

  need_root

  if [ "$os" = linux ]; then
    command -v systemctl >/dev/null 2>&1 || { echo "systemd is not available" >&2; exit 1; }
    curl -fsSL "$base/kanshi-agent.service" -o "$tmp/kanshi-agent.service"
    id kanshi-agent >/dev/null 2>&1 || $sudo useradd --system --home-dir /var/lib/kanshi-agent --shell /usr/sbin/nologin kanshi-agent
    $sudo install -d -o kanshi-agent -g kanshi-agent /var/lib/kanshi-agent
    printf 'KANSHI_CORE_ADDR=%s\nKANSHI_API_KEY=%s\nKANSHI_TLS=%s\nKANSHI_TLS_CA_FILE=%s\nKANSHI_TLS_SERVER_NAME=%s\nKANSHI_PROCESS_METRICS=%s\nKANSHI_PROCESS_TOP_N=%s\n' \
      "$KANSHI_CORE_ADDR" "$KANSHI_API_KEY" "${KANSHI_TLS:-}" "${KANSHI_TLS_CA_FILE:-}" "${KANSHI_TLS_SERVER_NAME:-}" "${KANSHI_PROCESS_METRICS:-}" "${KANSHI_PROCESS_TOP_N:-}" \
      | $sudo tee /etc/kanshi-agent.env >/dev/null
    $sudo chmod 0600 /etc/kanshi-agent.env
    sed "s#/usr/local/bin/kanshi-agent#$prefix/bin/kanshi-agent#" "$tmp/kanshi-agent.service" | $sudo tee /etc/systemd/system/kanshi-agent.service >/dev/null
    $sudo systemctl daemon-reload
    $sudo systemctl enable --now kanshi-agent
    echo "installed and started kanshi-agent $version as a systemd service"
  else
    # macOS: run at boot via a launchd daemon. The plist is written inline so no
    # extra release asset is needed. Keys are hex tokens, so no XML escaping.
    plist=/Library/LaunchDaemons/dev.kanshi.agent.plist
    printf '%s\n' \
      '<?xml version="1.0" encoding="UTF-8"?>' \
      '<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">' \
      '<plist version="1.0"><dict>' \
      '  <key>Label</key><string>dev.kanshi.agent</string>' \
      '  <key>ProgramArguments</key><array>' \
      "    <string>$prefix/bin/kanshi-agent</string>" \
      '  </array>' \
      '  <key>EnvironmentVariables</key><dict>' \
      "    <key>KANSHI_CORE_ADDR</key><string>$KANSHI_CORE_ADDR</string>" \
      "    <key>KANSHI_API_KEY</key><string>$KANSHI_API_KEY</string>" \
      "    <key>KANSHI_TLS</key><string>${KANSHI_TLS:-}</string>" \
      "    <key>KANSHI_TLS_CA_FILE</key><string>${KANSHI_TLS_CA_FILE:-}</string>" \
      "    <key>KANSHI_TLS_SERVER_NAME</key><string>${KANSHI_TLS_SERVER_NAME:-}</string>" \
      "    <key>KANSHI_PROCESS_METRICS</key><string>${KANSHI_PROCESS_METRICS:-}</string>" \
      "    <key>KANSHI_PROCESS_TOP_N</key><string>${KANSHI_PROCESS_TOP_N:-}</string>" \
      '  </dict>' \
      '  <key>RunAtLoad</key><true/>' \
      '  <key>KeepAlive</key><true/>' \
      '  <key>StandardOutPath</key><string>/var/log/kanshi-agent.log</string>' \
      '  <key>StandardErrorPath</key><string>/var/log/kanshi-agent.log</string>' \
      '</dict></plist>' \
      | $sudo tee "$plist" >/dev/null
    $sudo chown root:wheel "$plist"
    $sudo chmod 0600 "$plist"
    $sudo launchctl bootout system "$plist" 2>/dev/null || true
    $sudo launchctl bootstrap system "$plist"
    echo "installed and started kanshi-agent $version as a launchd service"
  fi
fi

echo "installed kanshi-agent $version to $prefix/bin/kanshi-agent"
