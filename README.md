<div align="center">

# 🔥 DEFLOT

**High-Performance Streaming Reconnaissance Engine for Bug Bounty Hunters**

[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

*Aggressively discover, deduplicate, and classify URLs from multiple passive sources*

[Features](#-features) • [Installation](#-installation) • [Quick Start](#-quick-start) • [Usage](#-usage) • [Documentation](#-documentation)

</div>

---

## 📋 Overview

**DEFLOT** is a production-grade reconnaissance tool designed for security professionals and bug bounty hunters. It excels at ingesting URLs from passive intelligence sources (Wayback Machine, VirusTotal, URLScan, OTX, GitHub), then normalizing, deduplicating, checking liveness, and automatically classifying them into high-value categories like secrets, configs, backups, and vulnerable parameters.

### 🎯 Core Philosophy

> **Streaming Pipeline + Low Memory Footprint + High Concurrency**

Unlike traditional tools that load everything into memory, DEFLOT uses a streaming architecture to process millions of URLs without exhausting RAM, making it ideal for large-scale reconnaissance.

---

## ✨ Features

<table>
<tr>
<td width="50%">

### 🚀 **Streaming Pipeline**
- Memory-efficient processing
- Handles millions of URLs
- Channel-based architecture
- Non-blocking I/O operations

### 🎯 **Smart Classification**
- **Secrets** (API keys, tokens, credentials)
- **Configs** (.env, .yml, .xml, config files)
- **Backups** (.bak, .old, .swp, .dump)
- **Parameters** (SQLi, XSS, IDOR vectors)
- **JavaScript** (with library filtering)
- **Documents** (PDFs, logs)
- **VCS Exposure** (.git, .svn)

</td>
<td width="50%">

### 🔌 **Multi-Source Integration**
- Wayback Machine
- VirusTotal API
- URLScan.io
- AlienVault OTX
- GitHub Code Search
- Local File Input

### ⚡ **Advanced Capabilities**
- Wildcard subdomain handling
- Thread-safe deduplication
- HTTP status filtering
- Batch target processing
- JSON output format
- JSSecretHunter integration
- Real-time Web UI

</td>
</tr>
</table>

---

## 📦 Installation

### Option 1: Install from Source (Recommended)

```bash
# Clone the repository
git clone https://github.com/elliot/deflot.git
cd deflot

# Build and install
go install
```

### Option 2: Build Locally

```bash
# Clone and build
git clone https://github.com/elliot/deflot.git
cd deflot
go build -o deflot

# Run from current directory
./deflot --help
```

### Option 3: Quick Install (if repository published)

```bash
go install github.com/elliot/deflot@latest
```

### Install Companion Tools (Optional)

DEFLOT integrates well with other security tools. Install them using:

```bash
chmod +x tools_install.sh
./tools_install.sh
```

This installs:
- **httpx** - HTTP probe
- **nuclei** - Vulnerability scanner
- **anew** - Deduplication utility
- **jq** - JSON processor
- **JSSecretHunter** - JavaScript secret scanner

---

## 🚀 Quick Start

### 1. Basic Domain Scan

```bash
deflot -d example.com
```

This will:
- Query all passive sources
- Deduplicate results
- Auto-save to `targets/example/`
- Display results in terminal

### 2. High-Performance Recon

```bash
deflot -d example.com \
  --wildcard \
  --sensitive-urls \
  --params \
  --js \
  -w 50 \
  -o ./results
```

### 3. Batch Processing

Create a target list:

```bash
cat > targets.txt << EOF
example.com
*.hackerone.com
github.com
EOF

deflot -t targets.txt
```

Results organized as `targets/example/`, `targets/hackerone/`, `targets/github/`

### 4. Integration Pipeline

Chain with other tools:

```bash
# Find live endpoints and scan for vulnerabilities
deflot -d example.com --json --stdout | \
  httpx -silent | \
  nuclei -t ~/nuclei-templates/
```

---

## 📖 Usage

### CLI Commands

#### Scan Mode (Primary)

```bash
deflot scan [flags]
# or simply
deflot [flags]
```

#### Configuration Management

```bash
# Initialize config file
deflot config --init

# Set API keys
deflot config --virustotal "YOUR_VT_KEY"
deflot config --urlscan "YOUR_URLSCAN_KEY"
deflot config --alienvault "YOUR_OTX_KEY"
deflot config --github "YOUR_GITHUB_TOKEN"
```

API keys are stored in `~/.deflot/config.yml`

#### Web Interface

```bash
deflot server
# Opens web UI at http://localhost:8080
```

---

### 🎛️ CLI Flags Reference

<details>
<summary><b>📍 Target Selection</b></summary>

| Flag | Short | Description | Example |
|------|-------|-------------|---------|
| `--domain` | `-d` | Target domain to scan | `-d example.com` |
| `--input` | `-i` | Input file of URLs/domains | `-i targets.txt` |
| `--target-list` | `-t` | Batch mode target list | `-t domains.txt` |
| `--wildcard` | | Enable wildcard subdomains | `--wildcard` |

</details>

<details>
<summary><b>🎯 Filters & Classification</b></summary>

| Flag | Description | Use Case |
|------|-------------|----------|
| `--sensitive-urls` | Filter for secrets, tokens, keys | Bug bounty reconnaissance |
| `--params` | Extract URLs with parameters | SQLi, XSS, IDOR hunting |
| `--js` | Filter JavaScript files | Code analysis, endpoints |
| `--exclude-libs` | Exclude common JS libraries | Remove jQuery, React, etc. |
| `--pdf` | Filter PDF documents | Information disclosure |
| `--log` | Filter log files | Sensitive data exposure |
| `--config` | Filter config files | Credential discovery |

</details>

<details>
<summary><b>⚡ Performance & Control</b></summary>

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--workers` | `-w` | 20 | Concurrent workers (1-100) |
| `--delay` | | 0ms | Delay between requests |
| `--timeout` | | 10s | HTTP request timeout |
| `--mc` | | | Match status codes (e.g., `200,403`) |

</details>

<details>
<summary><b>💾 Output Options</b></summary>

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Output directory (auto-created if not set) |
| `--json` | | JSON Lines output format |
| `--stdout` | | Stream to stdout (pipeable) |
| `--no-dedup` | | Disable deduplication (raw data) |

</details>

<details>
<summary><b>🔧 Advanced</b></summary>

| Flag | Description |
|------|-------------|
| `--js-scan` | Run JSSecretHunter on discovered JS files |
| `--sources` | Comma-separated source list (e.g., `wayback,virustotal`) |
| `--init-config` | Create default configuration file |

</details>

---

### 📂 Output Structure

When you specify an output directory (or use auto-save), DEFLOT organizes findings:

```
targets/example/
├── wayback_urls.txt                    # All discovered URLs
└── sensitiveurls/
    ├── secret_urls.txt                 # API keys, tokens, secrets
    ├── config_urls.txt                 # .env, .yml, .xml, .conf
    ├── database_backup_urls.txt        # .sql, .db, .dump
    ├── backup_exposure_urls.txt        # .bak, .old, .swp
    ├── api_specs_urls.txt              # swagger.json, openapi
    ├── parameter_urls.txt              # URLs with query params
    ├── js_urls.txt                     # JavaScript files
    ├── pdf_urls.txt                    # PDF documents
    ├── log_urls.txt                    # .log files
    └── vcs_exposure_urls.txt           # .git, .svn directories
```

---

## 🎓 Usage Examples

### Example 1: Complete Recon Workflow

```bash
# Step 1: Discover all URLs with advanced filters
deflot scan -d example.com \
  --wildcard \
  --sensitive-urls \
  --params \
  --js \
  --pdf \
  -w 50 \
  -o ./recon_results

# Step 2: Check for live endpoints
cat recon_results/wayback_urls.txt | httpx -silent -o live_urls.txt

# Step 3: Scan for vulnerabilities
cat live_urls.txt | nuclei -silent -severity critical,high
```

### Example 2: Secret Hunting

```bash
deflot -d target.com \
  --sensitive-urls \
  --js-scan \
  --no-dedup \
  -o ./secrets

# JS files are automatically scanned for hardcoded secrets
```

### Example 3: Parameter Discovery for Fuzzing

```bash
deflot -d example.com --params --stdout | \
  grep "=" | \
  awk -F'=' '{print $1"="}' | \
  sort -u > param_wordlist.txt
```

### Example 4: JSON Integration

```bash
# Get only config files as JSON
deflot -d example.com --json --stdout | \
  jq -c 'select(.category == "config")' | \
  jq -r '.url'
```

### Example 5: Multi-Target Bug Bounty

```bash
# Create your target list from your bug bounty programs
cat > my_targets.txt << EOF
# Program 1
example.com
*.example.com

# Program 2
target.io
api.target.io
EOF

# Batch scan all targets
deflot -t my_targets.txt --sensitive-urls --params -w 50
```

---

## 🏗️ Architecture

DEFLOT processes URLs through a strictly ordered pipeline:

```
┌─────────────┐
│   SOURCES   │  Passive collectors (Wayback, VT, URLScan, OTX, GitHub, File)
└──────┬──────┘
       │
       v
┌─────────────┐
│  NORMALIZE  │  Scheme enforcement, port stripping, canonicalization
└──────┬──────┘
       │
       v
┌─────────────┐
│ DEDUP GATE  │  Thread-safe duplicate removal (domain/wildcard scoped)
└──────┬──────┘
       │
       v
┌─────────────┐
│ STATUS GATE │  (Optional) Async HTTP HEAD/GET probes for liveness
└──────┬──────┘
       │
       v
┌─────────────┐
│   FILTERS   │  Regex-based classification (Secrets, JS, Params, Configs)
└──────┬──────┘
       │
       v
┌─────────────┐
│   OUTPUT    │  Buffered writing to categorized files + optional stdout
└─────────────┘
```

**Key Design Principles:**
- **Channels over slices**: Streaming prevents memory exhaustion
- **Goroutine workers**: High concurrency without blocking
- **Single-pass processing**: Each URL flows through once
- **Early filtering**: Remove duplicates/dead links before heavy processing

---

## 🔧 API Configuration

### Getting API Keys

1. **VirusTotal**: [https://www.virustotal.com/gui/my-apikey](https://www.virustotal.com/gui/my-apikey)
2. **URLScan.io**: [https://urlscan.io/user/profile](https://urlscan.io/user/profile)
3. **AlienVault OTX**: [https://otx.alienvault.com/api](https://otx.alienvault.com/api)
4. **GitHub**: [https://github.com/settings/tokens](https://github.com/settings/tokens) (needs `repo` scope for private repos)

### Configuration

```bash
# Method 1: Using config command
deflot config --virustotal "YOUR_KEY"
deflot config --urlscan "YOUR_KEY"
deflot config --alienvault "YOUR_KEY"
deflot config --github "YOUR_TOKEN"

# Method 2: Manual edit
nano ~/.deflot/config.yml
```

**Note:** DEFLOT works without API keys but will skip those sources. Wayback Machine and File input don't require authentication.

---

## 🔍 Troubleshooting

### Common Issues

<details>
<summary><b>No URLs found</b></summary>

- Ensure the domain has historical data in Wayback Machine
- Check if API keys are configured correctly
- Try `--sources wayback,virustotal` to debug specific sources
- Use `-w 5` to reduce concurrency if facing rate limits

</details>

<details>
<summary><b>Import errors after installation</b></summary>

```bash
# Re-download dependencies
go mod tidy
go mod download

# Rebuild
go build -o deflot
```

</details>

<details>
<summary><b>JSSecretHunter not found</b></summary>

```bash
# Install using the provided script
./tools_install.sh

# Or manually
mkdir -p ~/tools
git clone https://github.com/Cybertechhacks/JSSecretHunter.git ~/tools/JSSecretHunter
cd ~/tools/JSSecretHunter
python3 -m venv venv
./venv/bin/pip install -r requirements.txt
```

</details>

<details>
<summary><b>Permission denied on Linux</b></summary>

```bash
# Make binary executable
chmod +x deflot

# Or use go run
go run main.go scan -d example.com
```

</details>

---

## 🤝 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on:
- Reporting bugs
- Suggesting features
- Submitting pull requests
- Development setup

---

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

DEFLOT is built on the shoulders of giants:

- **ProjectDiscovery** - For inspiring security tool design
- **TomNomNom** - For the bug bounty methodology approach
- **Wayback Machine** - For preserving internet history
- **Go Community** - For excellent concurrency primitives

---

## ⚠️ Disclaimer

This tool is for **educational and authorized security testing purposes only**. Users are responsible for complying with applicable laws and obtaining proper authorization before scanning target networks. Unauthorized access to computer systems is illegal.

---

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/elliot/deflot/issues)
- **Discussions**: [GitHub Discussions](https://github.com/elliot/deflot/discussions)

---

<div align="center">

**Made with ❤️ for the Bug Bounty Community**

⭐ Star this repo if DEFLOT helped you find bugs!

</div>
