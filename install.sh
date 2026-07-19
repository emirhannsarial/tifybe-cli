#!/bin/sh
# Tifybe CLI installer — macOS & Linux
#   curl -fsSL https://tifybe.com/install.sh | sh
# Downloads the latest release binary for your platform, verifies its SHA-256
# checksum when available, and installs it as `tifybe`.
set -eu

REPO="emirhannsarial/tifybe-cli"
BASE="https://github.com/$REPO/releases/latest/download"

os=$(uname -s)
case "$os" in
  Linux)  os="linux" ;;
  Darwin) os="darwin" ;;
  *) echo "error: unsupported OS: $os (use install.ps1 on Windows)" >&2; exit 1 ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64|amd64)  arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) echo "error: unsupported architecture: $arch" >&2; exit 1 ;;
esac

asset="tifybe-$os-$arch"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading $asset (latest release)..."
curl -fsSL -o "$tmp/tifybe" "$BASE/$asset"

# Checksum verification — releases from v1.1.2 onward publish .sha256 files.
if curl -fsSL -o "$tmp/tifybe.sha256" "$BASE/$asset.sha256" 2>/dev/null; then
  expected=$(awk '{print $1}' "$tmp/tifybe.sha256")
  if command -v sha256sum >/dev/null 2>&1; then
    actual=$(sha256sum "$tmp/tifybe" | awk '{print $1}')
  else
    actual=$(shasum -a 256 "$tmp/tifybe" | awk '{print $1}')
  fi
  if [ "$expected" != "$actual" ]; then
    echo "error: checksum mismatch — aborting install" >&2
    exit 1
  fi
  echo "Checksum verified."
else
  echo "note: no checksum published for this release; skipping verification."
fi

chmod +x "$tmp/tifybe"

# Prefer /usr/local/bin when writable; otherwise ~/.local/bin (no sudo needed).
dest="/usr/local/bin"
if [ ! -w "$dest" ]; then
  if command -v sudo >/dev/null 2>&1 && [ -t 0 ]; then
    echo "Installing to $dest (sudo required)..."
    sudo mv "$tmp/tifybe" "$dest/tifybe"
  else
    dest="$HOME/.local/bin"
    mkdir -p "$dest"
    mv "$tmp/tifybe" "$dest/tifybe"
    case ":$PATH:" in
      *":$dest:"*) ;;
      *) echo "note: add $dest to your PATH:  export PATH=\"\$PATH:$dest\"" ;;
    esac
  fi
else
  mv "$tmp/tifybe" "$dest/tifybe"
fi

echo ""
echo "✓ tifybe installed to $dest/tifybe"
"$dest/tifybe" --version 2>/dev/null || true
echo ""
echo "Get started:  tifybe listen 8080"
