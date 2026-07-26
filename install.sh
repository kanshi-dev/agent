#!/bin/sh
set -eu

version=${KANSHI_VERSION:-v1.0.0}
prefix=${PREFIX:-/usr/local}
systemd=false
[ "${1:-}" = "--systemd" ] && systemd=true

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

if "$systemd"; then
  [ "$os" = linux ] || { echo "--systemd is supported only on Linux" >&2; exit 1; }
  : "${KANSHI_CORE_ADDR:?KANSHI_CORE_ADDR is required with --systemd}"
  : "${KANSHI_API_KEY:?KANSHI_API_KEY is required with --systemd}"
  command -v systemctl >/dev/null 2>&1 || { echo "systemd is not available" >&2; exit 1; }
  need_root
  curl -fsSL "$base/kanshi-agent.service" -o "$tmp/kanshi-agent.service"
  id kanshi-agent >/dev/null 2>&1 || $sudo useradd --system --home-dir /var/lib/kanshi-agent --shell /usr/sbin/nologin kanshi-agent
  $sudo install -d -o kanshi-agent -g kanshi-agent /var/lib/kanshi-agent
  printf 'KANSHI_CORE_ADDR=%s\nKANSHI_API_KEY=%s\n' "$KANSHI_CORE_ADDR" "$KANSHI_API_KEY" | $sudo tee /etc/kanshi-agent.env >/dev/null
  $sudo chmod 0600 /etc/kanshi-agent.env
  sed "s#/usr/local/bin/kanshi-agent#$prefix/bin/kanshi-agent#" "$tmp/kanshi-agent.service" | $sudo tee /etc/systemd/system/kanshi-agent.service >/dev/null
  $sudo systemctl daemon-reload
  $sudo systemctl enable --now kanshi-agent
fi

echo "installed kanshi-agent $version to $prefix/bin/kanshi-agent"
