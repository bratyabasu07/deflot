# Changelog

All notable changes to DEFLOT will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-02-03

### 🎉 Initial Release

DEFLOT v1.0.0 is a production-grade, streaming reconnaissance engine designed for bug bounty hunters and security professionals.

### Added

#### Core Features
- **Streaming Pipeline Architecture**: Low-memory footprint processing for massive target scans
- **Smart Deduplication Gate**: Thread-safe duplicate removal with wildcard subdomain support
- **HTTP Status Checker**: Async probing to filter dead links
- **Advanced Classification Engine**: Automatic categorization into Secrets, Configs, Backups, Parameters, JavaScript, PDFs, and Logs

#### Data Sources
- Wayback Machine integration
- VirusTotal API support
- URLScan.io integration
- AlienVault OTX connector
- GitHub code search
- Local file input mode

#### CLI Features
- Domain scanning mode (`-d`)
- File input mode (`-i`)
- Batch processing mode (`-t`) for multiple targets
- Wildcard subdomain handling (`--wildcard`)
- Custom output directories (`-o`)
- JSON Lines output format (`--json`)
- Stdout streaming (`--stdout`)
- Configurable concurrency (`-w`)
- HTTP status code filtering (`--mc`)

#### Filters
- Sensitive URL detection (API keys, secrets, tokens)
- Configuration file detection (.env, .yml, .xml)
- Backup exposure detection (.bak, .old, .swp)
- Parameter extraction (potential vulnerability points)
- JavaScript file classification (with library exclusion)
- PDF document discovery
- Log file detection
- VCS exposure detection (.git, .svn)

#### Advanced Features
- **JSSecretHunter Integration**: Automated secret scanning in JavaScript files
- **Web Interface**: Browser-based control panel (`deflot server`)
- **Auto-Save Mode**: Automatic result organization without explicit output flag
- **Target List Processing**: Batch scan multiple domains from a single file
- **Progress HUD**: Real-time statistics and source status display

#### Configuration
- API key management (`deflot config`)
- YAML-based configuration (`~/.deflot/config.yml`)
- Per-source enable/disable controls

#### Tools & Utilities
- Dependency installation script (`tools_install.sh`)
- Integration-ready with httpx, nuclei, anew, jq

### Performance
- Concurrent worker pool (default 20, configurable up to 50+)
- Buffered I/O for efficient file writing
- Streaming architecture prevents memory exhaustion
- Configurable delays and timeouts

### Documentation
- Comprehensive README with usage examples
- Full usage guide (`fullusage.md`)
- Web development guide (`WEB_DEV.md`)
- Contributing guidelines
- MIT License

---

## Future Roadmap

Planned features for upcoming releases:

- [ ] Enhanced passive DNS sources
- [ ] Screenshot capture integration
- [ ] Technology fingerprinting
- [ ] Custom filter rule engine
- [ ] Database output support
- [ ] GraphQL endpoint discovery
- [ ] Cloud storage detection (S3, Azure, GCS)
- [ ] Enhanced web UI with live controls

---

[1.0.0]: https://github.com/elliot/deflot/releases/tag/v1.0.0
