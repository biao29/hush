#!/usr/bin/env bash
#
# install.sh — Install hush binary from GitHub Releases.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/biao29/hush/main/scripts/install.sh | bash
#   curl -fsSL ... | HUSH_VERSION=v0.1.0 bash
#   curl -fsSL ... | HUSH_BIN_DIR=~/bin bash
#
set -euo pipefail

REPO="biao29/hush"
BIN_DIR="${HUSH_BIN_DIR:-${HOME}/.local/bin}"
VERSION="${HUSH_VERSION:-latest}"

# --- Helpers ----------------------------------------------------------------

info()  { printf "\033[0;32m✓\033[0m %s\n" "$1"; }
warn()  { printf "\033[0;33m!\033[0m %s\n" "$1"; }
error() { printf "\033[0;31m✗\033[0m %s\n" "$1" >&2; exit 1; }

need() {
    command -v "$1" >/dev/null 2>&1 || error "Required: $1 not found in PATH"
}

# --- Detect OS/Arch ---------------------------------------------------------

detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        *)       error "Unsupported OS: $(uname -s). Use install.ps1 for Windows." ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo "amd64" ;;
        aarch64|arm64)  echo "arm64" ;;
        *)              error "Unsupported architecture: $(uname -m)" ;;
    esac
}

# --- Resolve version --------------------------------------------------------

resolve_version() {
    if [ "$VERSION" = "latest" ]; then
        VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
            | grep '"tag_name"' | head -1 | cut -d'"' -f4)
        [ -n "$VERSION" ] || error "Could not determine latest version"
    fi
    echo "$VERSION"
}

# --- Main -------------------------------------------------------------------

need curl
need tar
if ! command -v sha256sum >/dev/null 2>&1 && ! command -v shasum >/dev/null 2>&1; then
    error "Required: sha256sum or shasum not found in PATH"
fi

OS=$(detect_os)
ARCH=$(detect_arch)
VERSION=$(resolve_version)

ARCHIVE="hush-${OS}-${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"
CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

info "Downloading hush ${VERSION} (${OS}/${ARCH})..."
curl -fsSL "$URL" -o "${TMPDIR}/${ARCHIVE}"
curl -fsSL "$CHECKSUMS_URL" -o "${TMPDIR}/checksums.txt"

# Verify checksum
info "Verifying checksum..."
cd "$TMPDIR"
if command -v sha256sum >/dev/null 2>&1; then
    grep "  ${ARCHIVE}$" checksums.txt | sha256sum -c --quiet -
else
    EXPECTED=$(grep "  ${ARCHIVE}$" checksums.txt | awk '{print $1}')
    ACTUAL=$(shasum -a 256 "${ARCHIVE}" | awk '{print $1}')
    [ "$EXPECTED" = "$ACTUAL" ] || error "Checksum mismatch: expected ${EXPECTED}, got ${ACTUAL}"
fi
info "Checksum verified"

# Extract and install
info "Installing to ${BIN_DIR}..."
tar xzf "${ARCHIVE}"
mkdir -p "${BIN_DIR}"
mv hush "${BIN_DIR}/hush"
chmod +x "${BIN_DIR}/hush"

info "hush ${VERSION} installed to ${BIN_DIR}/hush"

# Check PATH
if ! echo "$PATH" | tr ':' '\n' | grep -qx "${BIN_DIR}"; then
    echo ""
    warn "${BIN_DIR} is not in your PATH. Add it:"
    echo ""
    echo "  export PATH=\"${BIN_DIR}:\$PATH\""
    echo ""
    echo "  # Or add to your shell profile:"
    SHELL_NAME=$(basename "$SHELL")
    case "$SHELL_NAME" in
        zsh)  echo "  echo 'export PATH=\"${BIN_DIR}:\$PATH\"' >> ~/.zshrc" ;;
        bash) echo "  echo 'export PATH=\"${BIN_DIR}:\$PATH\"' >> ~/.bashrc" ;;
        fish) echo "  fish_add_path ${BIN_DIR}" ;;
        *)    echo "  echo 'export PATH=\"${BIN_DIR}:\$PATH\"' >> ~/.profile" ;;
    esac
fi
