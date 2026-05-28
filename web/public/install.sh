# Shipmate Installer
# Usage: curl -fsSL myshipmate.cc/install.sh | sh

set -e

REPO="emmogba/myshipmate"
GITHUB_API="https://api.github.com/repos/${REPO}/releases/latest"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="shipmate"
VERSION="${SHIPMATE_VERSION:-}"  # Allow override via env var

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

info() {
  printf "${GREEN}  ✓${NC} %s\n" "$1"
}

warn() {
  printf "${YELLOW}  ⚠${NC} %s\n" "$1"
}

error() {
  printf "${RED}  ✗${NC} %s\n" "$1" >&2
  exit 1
}

# Detect OS and architecture
detect_platform() {
  OS="$(uname -s)"
  ARCH="$(uname -m)"

  case "$OS" in
    Linux)  OS="linux" ;;
    Darwin) OS="darwin" ;;
    MINGW*|MSYS*|CYGWIN*) OS="windows" ;;
    *) error "Unsupported OS: $OS" ;;
  esac

  case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    armv7l) ARCH="arm" ;;
    *) error "Unsupported architecture: $ARCH" ;;
  esac

  echo "${OS}_${ARCH}"
}

# Main installation
main() {
  echo ""
  echo "  ╔═══════════════════════════════════════╗"
  echo "  ║       🚀 SHIPMATE INSTALLER           ║"
  echo "  ║   The Smart Deployer for Developers   ║"
  echo "  ╚═══════════════════════════════════════╝"
  echo ""

  PLATFORM=$(detect_platform)
  info "Detected platform: $PLATFORM"

  # Auto-detect latest version if not specified
  if [ -z "$VERSION" ]; then
    info "Checking latest version..."
    if command -v curl >/dev/null 2>&1; then
      VERSION=$(curl -fsSL "$GITHUB_API" 2>/dev/null | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
      VERSION=$(wget -qO- "$GITHUB_API" 2>/dev/null | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
    fi

    if [ -z "$VERSION" ]; then
      VERSION="v0.1.0"
      warn "Could not detect latest version, defaulting to $VERSION"
    else
      info "Latest version: $VERSION"
    fi
  else
    info "Installing specified version: $VERSION"
  fi

  TMPDIR=$(mktemp -d)
  trap "rm -rf $TMPDIR" EXIT

  BINARY_EXT=""
  if [ "$OS" = "windows" ]; then
    BINARY_EXT=".exe"
    BINARY_NAME="shipmate.exe"
  fi

  DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/shipmate_${PLATFORM}${BINARY_EXT}"

  info "Downloading Shipmate ${VERSION}..."

  # Try downloading pre-built binary
  DOWNLOAD_SUCCESS=false

  if command -v curl >/dev/null 2>&1; then
    if curl -fsSL -o "${TMPDIR}/${BINARY_NAME}" "$DOWNLOAD_URL" 2>/dev/null; then
      DOWNLOAD_SUCCESS=true
    fi
  elif command -v wget >/dev/null 2>&1; then
    if wget -q -O "${TMPDIR}/${BINARY_NAME}" "$DOWNLOAD_URL" 2>/dev/null; then
      DOWNLOAD_SUCCESS=true
    fi
  fi

  # Fallback: build from source
  if [ "$DOWNLOAD_SUCCESS" = false ]; then
    if command -v go >/dev/null 2>&1; then
      warn "Pre-built binary not available for $PLATFORM"
      info "Building from source with Go..."

      cd "$TMPDIR"
      git clone --depth 1 "https://github.com/${REPO}.git" . 2>/dev/null || {
        # If clone fails, the repo might not exist yet — that's ok during development
        warn "Could not clone repository."
        echo ""
        echo "  Shipmate is in early development."
        echo "  To install manually:"
        echo "    git clone https://github.com/${REPO}.git"
        echo "    cd cli"
        echo "    go build -o shipmate ."
        echo "    sudo mv shipmate /usr/local/bin/"
        echo ""
        exit 1
      }

      cd cli
      go mod tidy
      go build -o "$BINARY_NAME" .
      DOWNLOAD_SUCCESS=true
      TMPDIR_BIN="${TMPDIR}/cli/${BINARY_NAME}"
    else
      warn "Pre-built binary not available and Go is not installed."
      echo ""
      echo "  Options:"
      echo "    1. Install Go: https://go.dev/dl/"
      echo "       Then: go install github.com/${REPO}@latest"
      echo ""
      echo "    2. Build manually:"
      echo "       git clone https://github.com/${REPO}.git"
      echo "       cd cli && go build -o shipmate ."
      echo ""
      exit 1
    fi
  fi

  if [ -z "$TMPDIR_BIN" ]; then
    TMPDIR_BIN="${TMPDIR}/${BINARY_NAME}"
  fi

  chmod +x "$TMPDIR_BIN"

  # Install to target directory
  if [ -w "$INSTALL_DIR" ]; then
    mv "$TMPDIR_BIN" "${INSTALL_DIR}/${BINARY_NAME}"
  else
    info "Requesting sudo to install to ${INSTALL_DIR}..."
    sudo mv "$TMPDIR_BIN" "${INSTALL_DIR}/${BINARY_NAME}"
  fi

  info "Shipmate installed to ${INSTALL_DIR}/${BINARY_NAME}"
  echo ""
  echo "  Get started:"
  echo "    cd your-project"
  echo "    shipmate"
  echo ""
  echo "  Run 'shipmate --help' for all commands."
  echo ""
}

main "$@"
