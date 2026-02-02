# DEFLOT - Complete Usage Guide

## 📦 Installation

### Option 1: Install from Source (Recommended)

```bash
# Clone the repository
git clone https://github.com/bratyabasu07/deflot.git
cd deflot

# Install to your Go bin directory
go install

# Now you can run 'deflot' from anywhere (assuming $GOPATH/bin is in your PATH)
deflot --version
```

### Option 2: Build Locally

```bash
# Clone and build
git clone https://github.com/bratyabasu07/deflot.git
cd deflot
go build -o deflot

# Run from current directory
./deflot --help
```

### Option 3: Direct Install (if published on GitHub)

```bash
# Install latest release
go install github.com/bratyabasu07/deflot@latest

# Verify installation
deflot --version
```

### Installing Companion Tools

DEFLOT integrates well with other security tools:

```bash
# Make the install script executable
chmod +x tools_install.sh

# Install all companion tools
./tools_install.sh
```

This installs:
- **httpx** - Fast HTTP probe for checking liveness
- **nuclei** - Vulnerability scanner
- **anew** - Deduplication utility
- **jq** - JSON processor for piping output
- **JSSecretHunter** - JavaScript secret scanner

---

## 🔑 API Configuration

DEFLOT integrates with multiple threat intelligence sources. While it works without API keys (using Wayback Machine), adding keys dramatically increases coverage.

### Getting API Keys

1. **VirusTotal** (Free tier available)
   - Visit: https://www.virustotal.com/gui/my-apikey
   - Sign up and generate an API key
   - Free tier: 500 requests/day

2. **URLScan.io** (Free tier available)
   - Visit: https://urlscan.io/user/profile
   - Register and create an API key
   - Free tier: Unlimited public scans

3. **AlienVault OTX** (Free)
   - Visit: https://otx.alienvault.com/api
   - Create account and get API key
   - Completely free service

4. **GitHub** (Required for GitHub source)
   - Visit: https://github.com/settings/tokens
   - Generate personal access token
   - Required scopes: `repo` (for private repos), `public_repo` (for public)

### Configuring API Keys

#### Method 1: Using Config Command

```bash
# Initialize config file (creates ~/.deflot/config.yml)
deflot config --init

# Set individual API keys
deflot config --virustotal "YOUR_VT_API_KEY"
deflot config --urlscan "YOUR_URLSCAN_KEY"
deflot config --alienvault "YOUR_OTX_KEY"
deflot config --github "YOUR_GITHUB_TOKEN"

# Configure multiple at once
deflot config --virustotal "KEY" --github "TOKEN" --urlscan "KEY"
```

#### Method 2: Manual Configuration

Edit `~/.deflot/config.yml`:

```yaml
api_keys:
  virustotal: YOUR_VT_API_KEY
  urlscan: YOUR_URLSCAN_KEY
  alienvault: YOUR_OTX_KEY
  github: YOUR_GITHUB_TOKEN
```

**Note:** Keys are stored in your home directory at `~/.deflot/config.yml`

---

## 🚀 CLI Usage

### Basic Scanning

#### Simple Domain Scan

```bash
# Auto-save to targets/example/ directory
deflot scan -d example.com

# Or shorter syntax
deflot -d example.com
```

#### Custom Output Directory

```bash
deflot -d example.com -o /path/to/results
```

#### File Input Mode

```bash
# Scan URLs/domains from a file
deflot -i targets.txt

# With custom output
deflot -i targets.txt -o ./results
```

### Advanced Scanning

#### Complete Reconnaissance (All Features)

```bash
deflot scan -d example.com \
  --wildcard \              # Include all subdomains (*.example.com)
  --sensitive-urls \        # Filter for secrets, configs, backups
  --params \                # Extract URLs with parameters
  --js \                    # Filter JavaScript files
  --pdf \                   # Filter PDF documents
  --log \                   # Filter log files
  -w 50 \                   # Use 50 concurrent workers
  -o ./recon_results        # Save to custom directory
```

#### Wildcard Subdomain Enumeration

```bash
# Discover all subdomains from passive sources
deflot -d example.com --wildcard
```

#### Secret Hunting

```bash
# Focus on sensitive files and JavaScript secrets
deflot -d target.com \
  --sensitive-urls \
  --js-scan \
  --no-dedup \
  -o ./secrets
```

#### Parameter Discovery for Fuzzing

```bash
# Extract all URLs with parameters for SQLi/XSS testing
deflot -d example.com --params -o ./param_urls
```

### Batch Processing

#### Multiple Targets

Create a target list file:

```bash
cat > targets.txt << EOF
example.com
*.example1.com
example2.com
*.bugcrowd.com
EOF

# Scan all targets
deflot -t targets.txt --sensitive-urls
```

Results will be organized as:
- `targets/example/`
- `targets/example1/`
- `targets/example2/`

### Output Formats

#### JSON Output

```bash
# JSON Lines format (one JSON object per line)
deflot -d example.com --json -o ./results

# Stream to stdout (pipeable)
deflot -d example.com --json --stdout | jq .

# Extract specific categories
deflot -d example.com --json --stdout | jq -c 'select(.category == "secret")'
```

#### Standard Text Output

```bash
# Default: organized by category in text files
deflot -d example.com -o ./results
```

### Performance Tuning

#### Worker Concurrency

```bash
# Low concurrency (conservative, avoid rate limits)
deflot -d example.com -w 5

# Medium concurrency (default)
deflot -d example.com -w 20

# High concurrency (fast, use with caution)
deflot -d example.com -w 100
```

#### Request Delays

```bash
# Add delays to avoid rate limiting
deflot -d example.com --delay 100ms
```

#### HTTP Timeout

```bash
# Increase timeout for slow servers
deflot -d example.com --timeout 30s
```

### Filtering Options

#### Status Code Matching

```bash
# Only keep URLs that return specific status codes
deflot -d example.com --mc 200,403,401

# Useful for finding live endpoints or interesting responses
```

#### Deduplication Control

```bash
# Disable deduplication (keep all raw data)
deflot -d example.com --no-dedup
```

#### Source Selection

```bash
# Use specific sources only
deflot -d example.com --sources wayback,virustotal

# Available sources: wayback, virustotal, urlscan, alienvault, github
```

---

## 🌐 Web Interface Mode

DEFLOT includes a professional web interface for visual reconnaissance operations.

### Starting the Web Server

```bash
# Start with default settings (localhost:8080)
deflot server

# Custom port and address
deflot server --addr 127.0.0.1:3000

# Access from other machines (NOT RECOMMENDED for security)
deflot server --addr 0.0.0.0:8080
```

**Default URL:** http://127.0.0.1:8080

### Web UI Features

The web interface provides:

1. **Hacker Interface**
   - Dark mode with neon green accents
   - Live status indicators
   - Real-time output streaming

2. **Visual Controls**
   - Target input field
   - Weapon toggles (Wildcard, Sensitive, Params, JS, Exclude Libs)
   - Worker configuration
   - START RECON / STOP SCAN morphing button

3. **Live Output Stream**
   - Color-coded classifications:
     - **Green (System)**: System messages and status updates
     - **Red (Error)**: Errors and critical messages
     - **Yellow (JS)**: JavaScript files detected
     - **Red Bold (Secret)**: Secret/sensitive URLs found
     - **Cyan (Param)**: Parameter URLs
     - **Green (Result)**: Regular discovered URLs
   - Auto-scrolling
   - Clear and Copy buttons

4. **WebSocket Live Updates**
   - Real-time scan progress
   - Immediate feedback
   - No page reloads needed

### Using the Web Interface

1. **Start the server:**
   ```bash
   deflot server
   ```

2. **Open your browser:**
   - Navigate to http://127.0.0.1:8080

3. **Configure your scan:**
   - Enter target domain (e.g., `example.com`)
   - Toggle weapons (filters) as needed
   - Set worker count (default: 20)

4. **Run the scan:**
   - Click "START RECON"
   - Watch live output stream
   - Click "STOP SCAN" to abort if needed

5. **Review results:**
   - Use Clear button to reset output
   - Use Copy button to copy all results

### Web Mode Best Practices

- **Localhost Only**: Always use `127.0.0.1` for security
- **No Sensitive Data**: Web interface is for visual monitoring; results still save to disk
- **Performance**: Web mode adds minimal overhead, but CLI is faster for automation
- **Mobile Access**: Works on mobile browsers via localhost

---

## 📂 Output Structure

DEFLOT automatically organizes findings into categorized files:

```
targets/example/
├── wayback_urls.txt                    # All discovered URLs
└── sensitiveurls/
    ├── secret_urls.txt                 # API keys, tokens, credentials
    ├── config_urls.txt                 # .env, .yml, .xml, .conf files
    ├── database_backup_urls.txt        # .sql, .db, .dump files
    ├── backup_exposure_urls.txt        # .bak, .old, .swp files
    ├── api_specs_urls.txt              # swagger.json, openapi.yaml
    ├── parameter_urls.txt              # URLs with query parameters
    ├── js_urls.txt                     # JavaScript files
    ├── pdf_urls.txt                    # PDF documents
    ├── log_urls.txt                    # .log files
    └── vcs_exposure_urls.txt           # .git, .svn directories
```

### Category Descriptions

- **secret_urls.txt**: Contains patterns like `api_key=`, `token=`, `password=`, `secret=`
- **config_urls.txt**: Configuration files that may contain credentials
- **database_backup_urls.txt**: Database dumps that shouldn't be publicly accessible
- **backup_exposure_urls.txt**: Backup files often containing sensitive data
- **api_specs_urls.txt**: API documentation that reveals endpoints
- **parameter_urls.txt**: URLs with GET parameters (SQLi, XSS targets)
- **js_urls.txt**: JavaScript files (endpoints, secrets, logic)
- **pdf_urls.txt**: Documents that may contain sensitive information
- **log_urls.txt**: Log files often exposing internal data
- **vcs_exposure_urls.txt**: Version control directories (source code exposure)

---

## 🔗 Integration Workflows

### Example 1: Complete Bug Bounty Pipeline

```bash
# Step 1: Discover all URLs with comprehensive filters
deflot -d target.com \
  --wildcard \
  --sensitive-urls \
  --params \
  --js \
  --pdf \
  -w 50 \
  -o ./recon

# Step 2: Check for live endpoints
cat recon/wayback_urls.txt | httpx -silent -o live_urls.txt

# Step 3: Scan for vulnerabilities
cat live_urls.txt | nuclei -silent -severity critical,high

# Step 4: Scan JavaScript for secrets
cat recon/sensitiveurls/js_urls.txt | while read url; do
  curl -s "$url" | ~/tools/JSSecretHunter/scan.py
done
```

### Example 2: Secret Hunting Workflow

```bash
# Automated secret discovery
deflot -d target.com \
  --sensitive-urls \
  --js-scan \
  --no-dedup \
  -o ./secrets

# Review categorized findings
cat secrets/sensitiveurls/secret_urls.txt
cat secrets/sensitiveurls/config_urls.txt
```

### Example 3: Parameter Fuzzing Preparation

```bash
# Extract all parameter URLs
deflot -d example.com --params --stdout | \
  grep "=" | \
  awk -F'=' '{print $1"="}' | \
  sort -u > param_wordlist.txt

# Use with fuzzing tools
cat param_wordlist.txt | ffuf -w wordlist.txt -u "FUZZ"
```

### Example 4: JSON Processing

```bash
# Get only config files as JSON
deflot -d example.com --json --stdout | \
  jq -c 'select(.category == "config")' | \
  jq -r '.url' > config_urls.txt

# Filter by status code
deflot -d example.com --json --stdout | \
  jq -c 'select(.status_code == 200)' > live_200.json
```

### Example 5: Multi-Program Bug Bounty

```bash
# Create target list from bug bounty programs
cat > programs.txt << EOF
# Program 1
example.com
*.example.com

# Program 2
target.io
api.target.io

# Program 3
*.example1.com
EOF

# Batch scan all programs
deflot -t programs.txt \
  --sensitive-urls \
  --params \
  -w 50
```

---

## ⚡ Advanced Usage

### JSSecretHunter Integration

```bash
# Enable automatic JavaScript secret scanning
deflot -d target.com --js-scan

# This runs JSSecretHunter on every discovered JS file
```

### Wildcard Best Practices

```bash
# For domains with many subdomains
deflot -d example.com --wildcard -w 100

# Combine with sensitive URL filtering
deflot -d example.com --wildcard --sensitive-urls
```

### Library Exclusion

```bash
# Exclude common JS libraries (jQuery, React, etc.)
deflot -d target.com --js --exclude-libs

# Reduces noise in JavaScript findings
```

---

## 🐛 Troubleshooting

### No URLs Found

**Possible causes:**
- Domain has no historical data in Wayback Machine
- API keys not configured
- Rate limiting from sources

**Solutions:**
```bash
# Test specific sources
deflot -d example.com --sources wayback

# Reduce concurrency
deflot -d example.com -w 5

# Verify API configuration
cat ~/.deflot/config.yml
```

### Permission Denied

```bash
# Make binary executable
chmod +x deflot

# Or use go run
go run main.go scan -d example.com
```

### Import Errors

```bash
# Re-download dependencies
go mod tidy
go mod download

# Rebuild
go build -o deflot
```

### JSSecretHunter Not Found

```bash
# Install manually
mkdir -p ~/tools
git clone https://github.com/Cybertechhacks/JSSecretHunter.git ~/tools/JSSecretHunter
cd ~/tools/JSSecretHunter
python3 -m venv venv
./venv/bin/pip install -r requirements.txt
```

### Web Server Won't Start

```bash
# Check if port is already in use
lsof -i :8080

# Use different port
deflot server --addr 127.0.0.1:3000
```

---

## 📚 Command Reference

### Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--help` | `-h` | - | Show help |
| `--version` | | - | Show version |

### Scan Command

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--domain` | `-d` | - | Target domain |
| `--input` | `-i` | - | Input file of URLs |
| `--target-list` | `-t` | - | Batch target list |
| `--output` | `-o` | auto | Output directory |
| `--workers` | `-w` | 20 | Concurrent workers (1-100) |
| `--wildcard` | | false | Enable wildcard subdomains |
| `--sensitive-urls` | | false | Filter sensitive URLs |
| `--params` | | false | Extract parameter URLs |
| `--js` | | false | Filter JavaScript files |
| `--exclude-libs` | | false | Exclude JS libraries |
| `--pdf` | | false | Filter PDF documents |
| `--log` | | false | Filter log files |
| `--config` | | false | Filter config files |
| `--json` | | false | JSON output format |
| `--stdout` | | false | Stream to stdout |
| `--no-dedup` | | false | Disable deduplication |
| `--sources` | | all | Comma-separated sources |
| `--mc` | | - | Match status codes |
| `--delay` | | 0ms | Request delay |
| `--timeout` | | 10s | HTTP timeout |
| `--js-scan` | | false | Run JSSecretHunter |

### Config Command

| Flag | Description |
|------|-------------|
| `--init` | Create config file |
| `--virustotal` | Set VirusTotal API key |
| `--urlscan` | Set URLScan API key |
| `--alienvault` | Set AlienVault OTX key |
| `--github` | Set GitHub token |

### Server Command

| Flag | Default | Description |
|------|---------|-------------|
| `--addr` | 127.0.0.1:8080 | Server address |

---

## 💡 Pro Tips

1. **Start with Wayback Only**: Test domains with `--sources wayback` first (no API key needed)

2. **Use JSON for Automation**: Always use `--json --stdout` when piping to other tools

3. **Wildcard for Subdomains**: Enable `--wildcard` for comprehensive subdomain coverage

4. **Tune Workers**: Start with `-w 20`, increase to 50-100 for faster scans

5. **Save API Calls**: Use `--no-dedup` only when necessary; dedup saves API requests

6. **Combine Filters**: Stack flags like `--sensitive-urls --params --js` for targeted recon

7. **Web Mode for Demos**: Use `deflot server` for visual demonstrations and training

8. **Batch Processing**: Use `-t targets.txt` for managing multiple bug bounty programs

9. **Integration**: Combine with `httpx`, `nuclei`, `ffuf`, and `jq` for complete pipelines

10. **Status Codes**: Use `--mc 200,403` to focus on specific response types

---

**Made with ❤️ for the Bug Bounty Community**
