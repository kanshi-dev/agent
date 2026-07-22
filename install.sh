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
install -d "$prefix/bin"
install -m 0755 "$tmp/kanshi-agent" "$prefix/bin/kanshi-agent"

if "$systemd"; then
  [ "$os" = linux ] || { echo "--systemd is supported only on Linux" >&2; exit 1; }
  : "${KANSHI_CORE_ADDR:?KANSHI_CORE_ADDR is required with --systemd}"
  : "${KANSHI_API_KEY:?KANSHI_API_KEY is required with --systemd}"
  command -v systemctl >/dev/null 2>&1 || { echo "systemd is not available" >&2; exit 1; }
  curl -fsSL "$base/kanshi-agent.service" -o "$tmp/kanshi-agent.service"
  id kanshi-agent >/dev/null 2>&1 || useradd --system --home-dir /var/lib/kanshi-agent --shell /usr/sbin/nologin kanshi-agent
  install -d -o kanshi-agent -g kanshi-agent /var/lib/kanshi-agent
  printf 'KANSHI_CORE_ADDR=%s\nKANSHI_API_KEY=%s\n' "$KANSHI_CORE_ADDR" "$KANSHI_API_KEY" > /etc/kanshi-agent.env
  chmod 0600 /etc/kanshi-agent.env
  sed "s#/usr/local/bin/kanshi-agent#$prefix/bin/kanshi-agent#" "$tmp/kanshi-agent.service" > /etc/systemd/system/kanshi-agent.service
  systemctl daemon-reload
  systemctl enable --now kanshi-agent
fi

echo "installed kanshi-agent $version to $prefix/bin/kanshi-agent"
