#!/bin/bash
set -e

# Unbound UI - Bare Metal Installation Script
# Supports: Ubuntu/Debian, CentOS/RHEL, Alpine

echo "==================================="
echo "  Unbound UI - Installation Script"
echo "==================================="

# Check root
if [ "$EUID" -ne 0 ]; then
    echo "Error: Please run as root (sudo)"
    exit 1
fi

# Detect OS
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
else
    echo "Error: Cannot detect OS"
    exit 1
fi

echo "Detected OS: $OS"

# Install unbound if not present
install_unbound() {
    if command -v unbound &> /dev/null; then
        echo "Unbound is already installed"
        return
    fi

    echo "Installing Unbound..."
    case $OS in
        ubuntu|debian)
            apt-get update
            apt-get install -y unbound unbound-host
            ;;
        centos|rhel|fedora)
            dnf install -y unbound
            ;;
        alpine)
            apk add unbound
            ;;
        *)
            echo "Error: Unsupported OS: $OS"
            exit 1
            ;;
    esac
}

# Setup unbound
setup_unbound() {
    echo "Setting up Unbound..."

    # Create config directory
    mkdir -p /etc/unbound/unbound.conf.d

    # Setup remote control
    if [ ! -f /etc/unbound/unbound_control.key ]; then
        echo "Generating unbound-control keys..."
        unbound-control-setup
    fi

    # Enable and start unbound
    systemctl enable unbound
    systemctl start unbound || true
}

# Install unbound-ui binary
install_ui() {
    echo "Installing Unbound UI..."

    # Create directories
    mkdir -p /var/lib/unbound-ui/blocklist
    mkdir -p /etc/unbound-ui

    # Download or copy binary
    if [ -f "./bin/unbound-ui" ]; then
        cp ./bin/unbound-ui /usr/local/bin/unbound-ui
    else
        echo "Error: Binary not found. Please build first with 'make build'"
        exit 1
    fi

    chmod +x /usr/local/bin/unbound-ui

    # Create default config if not exists
    if [ ! -f /etc/unbound-ui/config.yaml ]; then
        cat > /etc/unbound-ui/config.yaml << 'EOF'
server:
  port: 8080
  host: "0.0.0.0"

unbound:
  config_path: "/etc/unbound/unbound.conf"
  control_path: "unbound-control"

auth:
  username: ""
  password_hash: ""
  jwt_secret: ""
  session_ttl: "24h"

blocklist:
  data_dir: "/var/lib/unbound-ui/blocklist"
  output_path: "/etc/unbound/unbound.conf.d/blocklist.conf"
  update_interval: "6h"
EOF
    fi

    # Create systemd service
    cat > /etc/systemd/system/unbound-ui.service << 'EOF'
[Unit]
Description=Unbound UI - Web Management Interface
After=network.target unbound.service
Requires=unbound.service

[Service]
Type=simple
ExecStart=/usr/local/bin/unbound-ui
WorkingDirectory=/etc/unbound-ui
Restart=always
RestartSec=5
User=root

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable unbound-ui
    systemctl start unbound-ui

    echo ""
    echo "==================================="
    echo "  Installation Complete!"
    echo "==================================="
    echo ""
    echo "Unbound UI is running on: http://$(hostname -I | awk '{print $1}'):8080"
    echo ""
    echo "On first visit, you'll be prompted to create an admin account."
    echo ""
    echo "Commands:"
    echo "  systemctl status unbound-ui    - Check status"
    echo "  systemctl restart unbound-ui   - Restart"
    echo "  journalctl -u unbound-ui -f    - View logs"
    echo ""
}

# Main
install_unbound
setup_unbound
install_ui
