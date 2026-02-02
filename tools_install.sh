#!/bin/bash
set -e

echo "[*] Installing DEFLOT Dependencies & Recommended Tools..."

# Check for Go
if ! command -v go &> /dev/null; then
    echo "[!] Go is not installed. Please install Go to use DEFLOT and related tools."
    exit 1
fi

echo "[+] Installing httpx (ProjectDiscovery)..."
go install -v github.com/projectdiscovery/httpx/cmd/httpx@latest

echo "[+] Installing nuclei (ProjectDiscovery)..."
go install -v github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest

echo "[+] Installing anew (TomNomNom)..."
go install -v github.com/tomnomnom/anew@latest


echo "[+] Installing jq..."
if command -v apt &> /dev/null; then
    sudo apt update && sudo apt install -y jq
elif command -v brew &> /dev/null; then
    brew install jq
else
    echo "[!] Could not install jq automatically. Please install it manually."
fi

echo "[*] checking for Python3..."
if ! command -v python3 &> /dev/null; then
    echo "[!] Python3 is required but not found. Please install Python3."
    exit 1
fi

echo "[+] Installing JSSecretHunter..."
TOOLS_DIR="$HOME/tools"
mkdir -p "$TOOLS_DIR"
if [ -d "$TOOLS_DIR/JSSecretHunter" ]; then
    echo "    JSSecretHunter already exists. Updating..."
    cd "$TOOLS_DIR/JSSecretHunter" && git pull
else
    git clone https://github.com/Cybertechhacks/JSSecretHunter.git "$TOOLS_DIR/JSSecretHunter"
fi

# Install python dependencies in a virtual environment
if [ -f "$TOOLS_DIR/JSSecretHunter/requirements.txt" ]; then
    echo "    Setting up virtual environment..."
    cd "$TOOLS_DIR/JSSecretHunter"
    python3 -m venv venv
    echo "    Installing python requirements in venv..."
    ./venv/bin/pip install -r requirements.txt
    echo "    JSSecretHunter installed. Use './venv/bin/python3 scanner_pro.py' to run."
fi

echo "[*] Installation Complete. Ensure $(go env GOPATH)/bin is in your PATH."
