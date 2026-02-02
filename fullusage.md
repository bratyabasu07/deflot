# DEFLOT - Full Usage Guide

## 1. Installation & Setup

### Running the Tool
You are currently running the tool using `./deflot`. This is because the binary is in your current directory but not in your system's global `PATH`.

To run it simply as `deflot`, you have two options:

**Option A: Install it (Recommended)**
```bash
go install .
# Now you can run 'deflot' from anywhere, assuming $(go env GOPATH)/bin is in your PATH.
```

**Option B: Add current dir to PATH (Temporary)**
```bash
export PATH=$PATH:.
# Now you can run 'deflot' from anywhere in this terminal session.
```

## 2. API Configuration

DEFLOT integrates with several threat intelligence APIs. You can configure them easily:

```bash
# Configure individual keys
./deflot config --virustotal "YOUR_VT_API_KEY"
./deflot config --urlscan "YOUR_URLSCAN_KEY"
./deflot config --alienvault "YOUR_OTX_KEY"
./deflot config --github "YOUR_GITHUB_TOKEN"

# Configure multiple at once
./deflot config --virustotal "KEY" --github "TOKEN"
```
These are saved to `~/.deflot/config.yml`.

## 3. Running Scans

### 🚀 Ultimate Recon (All-in-One)
Run everything to generate the full architecture output structure:
```bash
./deflot scan -d example.com --sensitive-urls --params --js --pdf --log --wildcard
```

### Basic Usage (Auto-Save Mode)
If you don't provide an output directory, DEFLOT will automatically create one with a timestamp (e.g., `deflot_results_20250125_120000`) and show results on screen.
```bash
./deflot scan -d example.com
```

### Advanced Recon
Combine multiple powerful flags for a deep scan:
```bash
./deflot scan \
  -d example.com \
  --wildcard \             # Handle all subdomains
  --no-dedup \             # Keep duplicates if you want raw data
  --params --js --pdf \    # Classify parameters, JS files, and PDFs
  -w 50 \                  # High concurrency (50 workers)
  -o ./bounty_data         # Custom output directory
```

### File Input Mode
Scan a list of URLs/Domains from a file:
```bash
./deflot scan -i targets.txt --json --stdout
```

### JSON Integration
Pipe structured data to other tools:
```bash
./deflot scan -d example.com --json --stdout | jq .
```

## 4. Output Structure
DEFLOT organizes your findings automatically:
```text
output_dir/
├── sensitiveurls/
│   ├── config_urls.txt           (.env, .yml, config.*)
│   ├── secret_urls.txt           (api keys, tokens)
│   ├── database_backup_urls.txt  (.sql, .db, .dump)
│   ├── backup_exposure_urls.txt  (.bak, .old)
│   ├── api_specs_urls.txt        (swagger.json, /api/)
│   └── ... (pdf, logs, vcs, etc)
└── wayback_urls.txt              (Unclassified URLs)
```
