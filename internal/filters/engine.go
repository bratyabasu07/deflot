package filters

import (
	"regexp"
	"strings"

	appCtx "github.com/elliot/deflot/internal/context"
)

// Categories
const (
	CatSecret   = "secret"
	CatConfig   = "config"
	CatBackup   = "backup"
	CatDatabase = "database"
	CatCloud    = "cloud"
	CatVCS      = "vcs"
	CatAPI      = "api"
	CatLog      = "log"
	CatArchive  = "archive"
	CatDoc      = "doc" // Covers pdf, doc, sheet for now, or split? Arch says "pdf_urls.txt / doc_urls.txt / sheet_urls.txt"
	CatPDF      = "pdf"
	CatSheet    = "sheet"
	CatJS       = "js"
	CatParam    = "param"
	CatNone     = "none"
)

// Engine handles URL classification.
type Engine struct {
	config appCtx.FilterConfig
	// Regex for expensive matches
	secretRegex *regexp.Regexp
}

// New creates a new filter engine.
func New(cfg appCtx.FilterConfig) *Engine {
	// Simple high-entropy/keyword regex for secrets
	// In production this would be much more complex.
	// Matching keys, tokens, auth, etc.
	r := regexp.MustCompile(`(?i)(api_key|access_token|secret|auth|password|passwd)`)

	return &Engine{
		config:      cfg,
		secretRegex: r,
	}
}

// Classify returns the highest priority category for a URL.
// Priority: Secret > Config > Backup > VCS > Cloud > Param > JS
func (e *Engine) Classify(urlStr string) string {
	// 1. Sensitive Files / Secrets (Highest Priority)
	if e.config.SensitiveUrls {
		if e.isSecret(urlStr) {
			return CatSecret
		}
		if e.isConfig(urlStr) {
			return CatConfig
		}
		if e.isBackup(urlStr) {
			return CatBackup
		}
		if e.isVCS(urlStr) {
			return CatVCS
		}
		if e.isDatabase(urlStr) {
			return CatDatabase
		}
		if e.isCloud(urlStr) {
			return CatCloud
		}
		if e.isAPI(urlStr) {
			return CatAPI
		}
		if e.isArchive(urlStr) {
			return CatArchive
		}
	}

	// Single Filters
	if e.config.Log && e.isLog(urlStr) {
		return CatLog
	}
	if e.config.PDF && e.isPDF(urlStr) {
		return CatPDF
	}
	// "Docs" and "Sheets" logic?
	// Arch lists doc_urls.txt and sheet_urls.txt.
	// We'll infer if "PDF" or generic document filter is enabled?
	// The flags in cmd/root.go are: --pdf, --log.
	// There is no --doc or --sheet flag.
	// But "SensitiveUrls" could cover documents?
	// Let's allow explicit classification if SensitiveUrls is on OR if we add implicit logic.
	// For now, I will check them under SensitiveUrls to ensure they get captured if that broader flag is on.
	if e.config.SensitiveUrls {
		if e.isDoc(urlStr) {
			return CatDoc
		}
		if e.isSheet(urlStr) {
			return CatSheet
		}
	}

	// 2. Parameters
	if e.config.Params && e.hasParams(urlStr) {
		return CatParam
	}

	// 3. JS
	if e.config.JS && e.isJS(urlStr) {
		if e.config.ExcludeLibs && e.isCommonLib(urlStr) {
			// If explicitly asked to exclude libs, we might return None here?
			// The requirement says "Filter classify URLs", "Priority-based single-category".
			// If it's a JS file but a common lib AND we actully exclude it, then it is noise.
			return CatNone
		}
		return CatJS
	}

	// If no filter matched (or filters disabled)
	// We might just return None.
	// But note: if filters are NOT enabled, should we classify?
	// The prompt implies filters are logic. If flag not set, we skip classification.

	return CatNone
}

func (e *Engine) isSecret(u string) bool {
	return e.secretRegex.MatchString(u)
}

func (e *Engine) isConfig(u string) bool {
	lower := strings.ToLower(u)
	return strings.Contains(lower, ".env") ||
		strings.Contains(lower, "config.") ||
		strings.Contains(lower, ".yml") ||
		strings.Contains(lower, ".xml") ||
		strings.Contains(lower, ".conf")
}

func (e *Engine) isBackup(u string) bool {
	lower := strings.ToLower(u)
	return strings.HasSuffix(lower, ".bak") ||
		strings.HasSuffix(lower, ".old") ||
		strings.HasSuffix(lower, ".swp") ||
		strings.Contains(lower, "backup")
}

func (e *Engine) isVCS(u string) bool {
	return strings.Contains(u, "/.git/") || strings.Contains(u, "/.svn/")
}

func (e *Engine) isCloud(u string) bool {
	return strings.Contains(u, "s3.amazonaws.com") ||
		strings.Contains(u, "blob.core.windows.net") ||
		strings.Contains(u, "storage.googleapis.com")
}

func (e *Engine) hasParams(u string) bool {
	return strings.Contains(u, "?") && strings.Contains(u, "=")
}

func (e *Engine) isJS(u string) bool {
	// Simple check, URL should usually be parsed but string check is faster for suffix
	// Be careful of query params: file.js?v=1
	// We want to check the path part.
	if strings.Contains(u, "?") {
		parts := strings.Split(u, "?")
		return strings.HasSuffix(strings.ToLower(parts[0]), ".js")
	}
	return strings.HasSuffix(strings.ToLower(u), ".js")
}

func (e *Engine) isCommonLib(u string) bool {
	lower := strings.ToLower(u)
	return strings.Contains(lower, "jquery") ||
		strings.Contains(lower, "bootstrap") ||
		strings.Contains(lower, "react") ||
		strings.Contains(lower, "vue")
}

func (e *Engine) isAPI(u string) bool {
	lower := strings.ToLower(u)
	return strings.Contains(lower, "/api/") ||
		strings.Contains(lower, "swagger") ||
		strings.Contains(lower, "openapi")
}

func (e *Engine) isLog(u string) bool {
	lower := strings.ToLower(u)
	return strings.HasSuffix(lower, ".log") || strings.Contains(lower, "error_log")
}

func (e *Engine) isArchive(u string) bool {
	lower := strings.ToLower(u)
	return strings.HasSuffix(lower, ".zip") || strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".rar")
}

func (e *Engine) isPDF(u string) bool {
	return strings.HasSuffix(strings.ToLower(u), ".pdf")
}

func (e *Engine) isDoc(u string) bool {
	lower := strings.ToLower(u)
	return strings.HasSuffix(lower, ".doc") || strings.HasSuffix(lower, ".docx") || strings.HasSuffix(lower, ".txt")
}

func (e *Engine) isSheet(u string) bool {
	lower := strings.ToLower(u)
	return strings.HasSuffix(lower, ".xls") || strings.HasSuffix(lower, ".xlsx") || strings.HasSuffix(lower, ".csv")
}

func (e *Engine) isDatabase(u string) bool {
	lower := strings.ToLower(u)
	return strings.HasSuffix(lower, ".sql") || strings.HasSuffix(lower, ".db") ||
		strings.HasSuffix(lower, ".dump") || strings.HasSuffix(lower, ".sqlite") ||
		strings.Contains(lower, "mysqldump")
}
