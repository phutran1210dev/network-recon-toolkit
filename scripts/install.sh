#!/bin/bash
# Network Recon Toolkit Installation Script

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO_OWNER="phutran1210dev"
REPO_NAME="network-recon-toolkit"
BINARY_NAME="netrecon"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/netrecon"
GITHUB_API_URL="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"

# Utility functions
log() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Detect OS and architecture
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case $os in
        linux)
            OS="linux"
            ;;
        darwin)
            OS="darwin"
            ;;
        *)
            error "Unsupported operating system: $os"
            ;;
    esac
    
    case $arch in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        armv7l)
            ARCH="armv7"
            ;;
        *)
            error "Unsupported architecture: $arch"
            ;;
    esac
    
    PLATFORM="${OS}_${ARCH}"
    log "Detected platform: $PLATFORM"
}

# Check if running as root for system installation
check_permissions() {
    if [[ $EUID -eq 0 ]]; then
        SYSTEM_INSTALL=true
        INSTALL_DIR="/usr/local/bin"
        CONFIG_DIR="/etc/netrecon"
    else
        SYSTEM_INSTALL=false
        INSTALL_DIR="$HOME/.local/bin"
        CONFIG_DIR="$HOME/.config/netrecon"
        
        # Create local bin directory if it doesn't exist
        mkdir -p "$INSTALL_DIR"
        
        # Add to PATH if not already there
        if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
            warn "Add $INSTALL_DIR to your PATH for easy access:"
            echo "echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc"
            echo "source ~/.bashrc"
        fi
    fi
    
    log "Installation directory: $INSTALL_DIR"
    log "Configuration directory: $CONFIG_DIR"
}

# Get latest release information
get_latest_release() {
    log "Fetching latest release information..."
    
    if command -v curl &> /dev/null; then
        RELEASE_INFO=$(curl -s "$GITHUB_API_URL")
    elif command -v wget &> /dev/null; then
        RELEASE_INFO=$(wget -qO- "$GITHUB_API_URL")
    else
        error "Neither curl nor wget is available. Please install one of them."
    fi
    
    # Extract version and download URL
    VERSION=$(echo "$RELEASE_INFO" | grep '"tag_name":' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')
    
    if [[ "$ARCH" == "amd64" ]]; then
        ARCH_PATTERN="x86_64"
    else
        ARCH_PATTERN="$ARCH"
    fi
    
    DOWNLOAD_URL=$(echo "$RELEASE_INFO" | grep '"browser_download_url":' | grep "${OS}" | grep "${ARCH_PATTERN}" | head -1 | sed -E 's/.*"browser_download_url": "([^"]+)".*/\1/')
    
    if [[ -z "$VERSION" ]] || [[ -z "$DOWNLOAD_URL" ]]; then
        error "Could not find release information for platform $PLATFORM"
    fi
    
    log "Latest version: $VERSION"
    log "Download URL: $DOWNLOAD_URL"
}

# Download and install binary
install_binary() {
    log "Downloading Network Recon Toolkit $VERSION..."
    
    TEMP_DIR=$(mktemp -d)
    ARCHIVE_NAME="$(basename "$DOWNLOAD_URL")"
    
    cd "$TEMP_DIR"
    
    if command -v curl &> /dev/null; then
        curl -L -o "$ARCHIVE_NAME" "$DOWNLOAD_URL"
    else
        wget -O "$ARCHIVE_NAME" "$DOWNLOAD_URL"
    fi
    
    log "Extracting archive..."
    if [[ "$ARCHIVE_NAME" == *.tar.gz ]]; then
        tar -xzf "$ARCHIVE_NAME"
    elif [[ "$ARCHIVE_NAME" == *.zip ]]; then
        unzip -q "$ARCHIVE_NAME"
    else
        error "Unsupported archive format: $ARCHIVE_NAME"
    fi
    
    # Find the binary
    BINARY_PATH=$(find . -name "$BINARY_NAME" -type f | head -1)
    if [[ -z "$BINARY_PATH" ]]; then
        error "Binary not found in archive"
    fi
    
    log "Installing binary to $INSTALL_DIR..."
    if [[ "$SYSTEM_INSTALL" == true ]]; then
        sudo install -m 755 "$BINARY_PATH" "$INSTALL_DIR/"
    else
        install -m 755 "$BINARY_PATH" "$INSTALL_DIR/"
    fi
    
    # Clean up
    rm -rf "$TEMP_DIR"
    
    log "Binary installed successfully!"
}

# Install configuration files
install_config() {
    log "Setting up configuration..."
    
    if [[ "$SYSTEM_INSTALL" == true ]]; then
        sudo mkdir -p "$CONFIG_DIR"
        
        # Create default config if it doesn't exist
        if [[ ! -f "$CONFIG_DIR/config.yaml" ]]; then
            sudo tee "$CONFIG_DIR/config.yaml" > /dev/null <<EOF
database:
  host: localhost
  port: 5432
  user: netrecon
  password: netrecon_password
  dbname: netrecon
  sslmode: disable

logging:
  level: info
  format: text
  file: ""

scanner:
  default_timeout: 300
  max_threads: 1000
  default_ports: "1-1000"
  presets:
    quick:
      scanner: nmap
      ports: "22,23,25,53,80,110,443,993,995"
      arguments: "-sS"
      timing: "4"
    comprehensive:
      scanner: nmap
      ports: "1-65535"
      arguments: "-sS -sV -O -A"
      timing: "4"
    fast:
      scanner: masscan
      ports: "1-1000"
      arguments: ""
      timing: "4"
    web:
      scanner: nmap
      ports: "80,443,8080,8443,8000,8888"
      arguments: "-sS -sV --script http-enum"
      timing: "4"

server:
  host: localhost
  port: 8080
EOF
            log "Default configuration created at $CONFIG_DIR/config.yaml"
        fi
    else
        mkdir -p "$CONFIG_DIR"
        
        if [[ ! -f "$CONFIG_DIR/config.yaml" ]]; then
            # Create user config (same content as above)
            cat > "$CONFIG_DIR/config.yaml" <<EOF
database:
  host: localhost
  port: 5432
  user: netrecon
  password: netrecon_password
  dbname: netrecon
  sslmode: disable

logging:
  level: info
  format: text
  file: ""

scanner:
  default_timeout: 300
  max_threads: 1000
  default_ports: "1-1000"
  presets:
    quick:
      scanner: nmap
      ports: "22,23,25,53,80,110,443,993,995"
      arguments: "-sS"
      timing: "4"
    comprehensive:
      scanner: nmap
      ports: "1-65535"
      arguments: "-sS -sV -O -A"
      timing: "4"
    fast:
      scanner: masscan
      ports: "1-1000"
      arguments: ""
      timing: "4"
    web:
      scanner: nmap
      ports: "80,443,8080,8443,8000,8888"
      arguments: "-sS -sV --script http-enum"
      timing: "4"

server:
  host: localhost
  port: 8080
EOF
            log "Configuration created at $CONFIG_DIR/config.yaml"
        fi
    fi
}

# Check dependencies
check_dependencies() {
    log "Checking dependencies..."
    
    # Check for nmap
    if ! command -v nmap &> /dev/null; then
        warn "nmap not found. Install it for full functionality:"
        case $OS in
            linux)
                echo "  Ubuntu/Debian: sudo apt-get install nmap"
                echo "  RHEL/CentOS: sudo yum install nmap"
                echo "  Arch: sudo pacman -S nmap"
                ;;
            darwin)
                echo "  macOS: brew install nmap"
                ;;
        esac
    else
        log "nmap found: $(nmap --version | head -1)"
    fi
    
    # Check for masscan
    if ! command -v masscan &> /dev/null; then
        warn "masscan not found. Install it for high-speed scanning:"
        case $OS in
            linux)
                echo "  Ubuntu/Debian: sudo apt-get install masscan"
                echo "  Build from source: https://github.com/robertdavidgraham/masscan"
                ;;
            darwin)
                echo "  macOS: brew install masscan"
                ;;
        esac
    else
        log "masscan found: $(masscan --version 2>&1 | head -1)"
    fi
}

# Verify installation
verify_installation() {
    log "Verifying installation..."
    
    if "$INSTALL_DIR/$BINARY_NAME" version &> /dev/null; then
        log "Installation successful!"
        echo ""
        echo -e "${BLUE}Network Recon Toolkit${NC} has been installed successfully!"
        echo ""
        echo "Version information:"
        "$INSTALL_DIR/$BINARY_NAME" version
        echo ""
        echo "Quick start:"
        echo "  $BINARY_NAME --help"
        echo "  $BINARY_NAME scan --help"
        echo ""
        echo "Configuration file: $CONFIG_DIR/config.yaml"
        echo ""
        echo "For more information, visit: https://github.com/$REPO_OWNER/$REPO_NAME"
    else
        error "Installation verification failed"
    fi
}

# Main installation process
main() {
    echo -e "${BLUE}"
    echo "╔══════════════════════════════════════════════════════════════════════════════╗"
    echo "║                      Network Recon Toolkit Installer                        ║"
    echo "╚══════════════════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
    
    detect_platform
    check_permissions
    get_latest_release
    install_binary
    install_config
    check_dependencies
    verify_installation
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Network Recon Toolkit Installation Script"
        echo ""
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "This script will download and install the latest version of Network Recon Toolkit"
        echo "from GitHub releases."
        echo ""
        echo "Options:"
        echo "  --help, -h    Show this help message"
        echo ""
        echo "The script will:"
        echo "  - Detect your OS and architecture"
        echo "  - Download the appropriate binary"
        echo "  - Install to /usr/local/bin (with sudo) or ~/.local/bin (user install)"
        echo "  - Set up configuration files"
        echo "  - Check for dependencies (nmap, masscan)"
        echo ""
        echo "Examples:"
        echo "  curl -sSL https://raw.githubusercontent.com/$REPO_OWNER/$REPO_NAME/master/scripts/install.sh | bash"
        echo "  wget -qO- https://raw.githubusercontent.com/$REPO_OWNER/$REPO_NAME/master/scripts/install.sh | bash"
        exit 0
        ;;
    *)
        main
        ;;
esac