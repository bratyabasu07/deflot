package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bratyabasu07/deflot/internal/config"
	appCtx "github.com/bratyabasu07/deflot/internal/context"
	"github.com/bratyabasu07/deflot/internal/dedup"
	"github.com/bratyabasu07/deflot/internal/filters"
	"github.com/bratyabasu07/deflot/internal/integrations/jssecrethunter"
	"github.com/bratyabasu07/deflot/internal/output"
	"github.com/bratyabasu07/deflot/internal/pipeline"
	"github.com/bratyabasu07/deflot/internal/sources"
	"github.com/bratyabasu07/deflot/internal/status"
	"github.com/bratyabasu07/deflot/internal/summary"
	"github.com/bratyabasu07/deflot/internal/targetlist"
	"github.com/bratyabasu07/deflot/internal/ui"

	"github.com/spf13/cobra"
)

var (
	// Required flags
	domainFlag     string
	inputFlag      string
	targetListFlag string

	// Output flags
	outputFlag string

	// Performance flags
	workersFlag int
	delayFlag   int
	timeoutFlag int

	// Filter flags
	sensitiveUrlsFlag bool
	paramsFlag        bool
	jsFlag            bool
	excludeLibsFlag   bool
	pdfFlag           bool
	logFlag           bool
	configFilterFlag  bool

	// Scanners
	jsScanFlag bool

	// Advanced flags
	wildcardFlag bool
	noDedupFlag  bool
	mcFlag       string

	sourcesFlag    string
	initConfigFlag bool

	// Output format flags
	jsonFlag   bool
	stdoutFlag bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "deflot",
	Version: "v1.0.0",
	Short:   "High-performance streaming bug bounty reconnaissance tool",
	Long: ui.Banner + `
DEFLOT is a production-grade, streaming reconnaissance engine designed for security professionals.
It aggressively discovers, deduplicates, and classifies URLs from multiple passive sources.

Features:
  - Streaming Pipeline: Low memory footprint even for massive targets
  - Smart Gates: Deduplication and HTTP Status checking
  - Classification: Automatically tags Secrets, Configs, Backups, Params, JS
  - Modular Sources: VirusTotal, URLScan, OTX, AlienVault, GitHub, Wayback Machine
`,
	Example: `  # Basic scan of a domain
  deflot -d example.com -o ./results

  # Batch scan multiple targets from a file
  deflot -t targets.txt

  # Using the scan subcommand
  deflot scan -d example.com --json --stdout
`,
	Run: runScan, // Refactored to separate function
}

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Run the reconnaissance scanner",
	Run:   runScan,
}

func runScan(cmd *cobra.Command, args []string) {
	// 1. Handle --init-config
	if initConfigFlag {
		config.GenerateDefaultConfig()
		return
	}

	// 2. Validation
	if domainFlag == "" && inputFlag == "" && targetListFlag == "" {
		cmd.Help()
		fmt.Println("\n[!] Error: You must provide a target domain (-d), input file (-i), or target list (-t).")
		os.Exit(1)
	}

	// Check for mutually exclusive flags
	if (domainFlag != "" && targetListFlag != "") || (inputFlag != "" && targetListFlag != "") {
		fmt.Println("[!] Error: -t/--target-list cannot be used with -d or -i flags.")
		os.Exit(1)
	}

	// Batch mode handling
	if targetListFlag != "" {
		runBatchScan(cmd, args)
		return
	}

	// 3. Start Cinematic Intro
	// Check isJSON/isStdout from CLI flags before auto-save defaults kick in.
	stopIntro := ui.StartIntro(jsonFlag, stdoutFlag)
	// Ensure we stop it on exit/panic safety, though we call it explicitly later
	defer stopIntro()

	// 3. Auto-Save Logic
	// If output (-o) is NOT provided, structure output in 'targets/<name>'
	if outputFlag == "" {
		var targetName string
		if domainFlag != "" {
			// Extract primary name (e.g., "example" from "example.com")
			parts := strings.Split(domainFlag, ".")
			if len(parts) > 0 {
				targetName = parts[0]
			} else {
				targetName = domainFlag
			}
		} else if inputFlag != "" {
			// Extract filename without extension
			base := filepath.Base(inputFlag)
			ext := filepath.Ext(base)
			targetName = strings.TrimSuffix(base, ext)
		} else {
			targetName = "unknown"
		}

		// Sanitize target name (basic)
		targetName = strings.Map(func(r rune) rune {
			if r == '/' || r == '\\' || r == ':' {
				return '_'
			}
			return r
		}, targetName)

		outputFlag = filepath.Join("targets", targetName)
		stdoutFlag = true // Ensure user sees output since they didn't explicitly ask for file only
		fmt.Printf("[*] Auto-Save Enabled: Results saving to %s\n", outputFlag)
	}

	// 4. Initialize Context
	if jsScanFlag {
		jsFlag = true
		fmt.Println("[*] Enabling JS Filter for Scanner...")
	}

	filterCfg := appCtx.FilterConfig{
		SensitiveUrls: sensitiveUrlsFlag,
		Params:        paramsFlag,
		JS:            jsFlag,
		ExcludeLibs:   excludeLibsFlag,
		PDF:           pdfFlag,
		Log:           logFlag,
		Config:        configFilterFlag,
	}

	appContext, err := appCtx.New(
		domainFlag, inputFlag, wildcardFlag, outputFlag, sourcesFlag,
		workersFlag, delayFlag, timeoutFlag, noDedupFlag, mcFlag,
		jsonFlag, stdoutFlag, filterCfg,
	)
	if err != nil {
		fmt.Printf("[!] Initialization Error: %v\n", err)
		os.Exit(1)
	}

	// 4. Initialize Components
	cfg := config.Config{ApiKeys: config.GetAPIKeys()}

	// Managers
	sourceMgr := sources.NewManager(appContext, cfg)

	// API Sources (always registered, Manager handles enable/disable/keys)
	sourceMgr.Register(sources.NewWayback(appContext.Domain))
	sourceMgr.Register(sources.NewVirusTotal(appContext.Domain, cfg))
	sourceMgr.Register(sources.NewURLScan(appContext.Domain, cfg))
	sourceMgr.Register(sources.NewAlienVault(appContext.Domain, cfg))
	sourceMgr.Register(sources.NewGitHub(appContext.Domain, cfg))

	// Local Source
	if inputFlag != "" {
		sourceMgr.Register(sources.NewFileSource(inputFlag))
	}

	stats := summary.New()

	// Utility Engines
	deduplicator := dedup.New(appContext.Domain, appContext.Wildcard, appContext.NoDedup)
	checker := status.New(appContext.Timeout, appContext.Match)
	filterEngine := filters.New(appContext.Filters)
	flasher := ui.NewFlasher(jsonFlag, stdoutFlag)

	writer, err := output.New(appContext)
	if err != nil {
		fmt.Printf("[!] Output Error: %v\n", err)
		os.Exit(1)
	}
	defer writer.Close()

	// Scanners
	jsScanner := jssecrethunter.New(jsScanFlag)
	if jsScanFlag && !jsScanner.IsAvailable() {
		fmt.Println("[!] Warning: JSSecretHunter not found. Run tools_install.sh to install.")
	}

	pipe := pipeline.New(appContext, deduplicator, checker, filterEngine, writer, stats, flasher.Notify, jsScanner)

	// 5. Execution Flow
	ctx := context.Background()

	fmt.Printf("[*] Target: %s\n", appContext.Domain)
	if appContext.OutputDir != "" {
		fmt.Printf("[*] Output: %s\n", appContext.OutputDir)
	}

	// Start Sources
	// This returns a channel of raw URLs
	stopIntro() // Stop animation before massive output

	// Start HUD
	stopHUD := ui.StartHUD(ctx, stats, sourceMgr, jsonFlag, stdoutFlag)
	defer stopHUD()

	rawChan := sourceMgr.StartAll(ctx)

	// Start Pipeline
	done := pipe.Start(ctx, rawChan)

	<-done
	stats.PrintReport()
	ui.PrintOutro(jsonFlag, stdoutFlag)
}

// runBatchScan processes multiple targets from a target list file.
func runBatchScan(cmd *cobra.Command, args []string) {
	// Parse target list
	targets, err := targetlist.ParseTargetList(targetListFlag)
	if err != nil {
		fmt.Printf("[!] Error parsing target list: %v\\n", err)
		os.Exit(1)
	}

	totalTargets := len(targets)
	fmt.Printf("\\n[*] Batch Mode: Processing %d targets from %s\\n\\n", totalTargets, targetListFlag)

	// Process each target sequentially
	for idx, target := range targets {
		targetNum := idx + 1
		fmt.Printf("\\n" + strings.Repeat("=", 70) + "\\n")
		fmt.Printf("[Target %d/%d] Starting scan: %s\\n", targetNum, totalTargets, target)
		fmt.Printf(strings.Repeat("=", 70) + "\\n\\n")

		// Temporarily set the domain flag for this specific target
		originalDomain := domainFlag
		originalOutput := outputFlag
		domainFlag = target

		// Set per-target output directory
		if originalOutput == "" {
			// Extract target name for directory
			parts := strings.Split(target, ".")
			var targetName string
			if strings.HasPrefix(target, "*.") {
				if len(parts) > 1 {
					targetName = parts[1]
				} else {
					targetName = "wildcard"
				}
			} else if len(parts) > 0 {
				targetName = parts[0]
			} else {
				targetName = target
			}
			targetName = strings.Map(func(r rune) rune {
				if r == '/' || r == '\\' || r == ':' || r == '*' {
					return '_'
				}
				return r
			}, targetName)
			outputFlag = filepath.Join("targets", targetName)
		} else {
			parts := strings.Split(target, ".")
			var targetName string
			if strings.HasPrefix(target, "*.") {
				if len(parts) > 1 {
					targetName = parts[1]
				} else {
					targetName = "wildcard"
				}
			} else if len(parts) > 0 {
				targetName = parts[0]
			} else {
				targetName = target
			}
			targetName = strings.Map(func(r rune) rune {
				if r == '/' || r == '\\' || r == ':' || r == '*' {
					return '_'
				}
				return r
			}, targetName)
			outputFlag = filepath.Join(originalOutput, targetName)
		}

		// Execute the scan for this target
		runSingleTargetScan()

		// Restore original flags
		domainFlag = originalDomain
		outputFlag = originalOutput

		fmt.Printf("\\n[✓] Completed target %d/%d: %s\\n", targetNum, totalTargets, target)
	}

	fmt.Printf("\\n" + strings.Repeat("=", 70) + "\\n")
	fmt.Printf("[✓] Batch scan complete! Processed %d targets.\\n", totalTargets)
	fmt.Printf(strings.Repeat("=", 70) + "\\n\\n")
}

// runSingleTargetScan executes the scan logic for a single target.
func runSingleTargetScan() {
	if outputFlag == "" {
		var targetName string
		if domainFlag != "" {
			parts := strings.Split(domainFlag, ".")
			if len(parts) > 0 {
				targetName = parts[0]
			} else {
				targetName = domainFlag
			}
		} else if inputFlag != "" {
			base := filepath.Base(inputFlag)
			ext := filepath.Ext(base)
			targetName = strings.TrimSuffix(base, ext)
		} else {
			targetName = "unknown"
		}

		targetName = strings.Map(func(r rune) rune {
			if r == '/' || r == '\\' || r == ':' {
				return '_'
			}
			return r
		}, targetName)

		outputFlag = filepath.Join("targets", targetName)
		stdoutFlag = true
		fmt.Printf("[*] Auto-Save Enabled: Results saving to %s\\n", outputFlag)
	}

	if jsScanFlag {
		jsFlag = true
		fmt.Println("[*] Enabling JS Filter for Scanner...")
	}

	filterCfg := appCtx.FilterConfig{
		SensitiveUrls: sensitiveUrlsFlag,
		Params:        paramsFlag,
		JS:            jsFlag,
		ExcludeLibs:   excludeLibsFlag,
		PDF:           pdfFlag,
		Log:           logFlag,
		Config:        configFilterFlag,
	}

	appContext, err := appCtx.New(
		domainFlag, inputFlag, wildcardFlag, outputFlag, sourcesFlag,
		workersFlag, delayFlag, timeoutFlag, noDedupFlag, mcFlag,
		jsonFlag, stdoutFlag, filterCfg,
	)
	if err != nil {
		fmt.Printf("[!] Initialization Error: %v\\n", err)
		os.Exit(1)
	}

	cfg := config.Config{ApiKeys: config.GetAPIKeys()}
	sourceMgr := sources.NewManager(appContext, cfg)

	sourceMgr.Register(sources.NewWayback(appContext.Domain))
	sourceMgr.Register(sources.NewVirusTotal(appContext.Domain, cfg))
	sourceMgr.Register(sources.NewURLScan(appContext.Domain, cfg))
	sourceMgr.Register(sources.NewAlienVault(appContext.Domain, cfg))
	sourceMgr.Register(sources.NewGitHub(appContext.Domain, cfg))

	if inputFlag != "" {
		sourceMgr.Register(sources.NewFileSource(inputFlag))
	}

	stats := summary.New()
	deduplicator := dedup.New(appContext.Domain, appContext.Wildcard, appContext.NoDedup)
	checker := status.New(appContext.Timeout, appContext.Match)
	filterEngine := filters.New(appContext.Filters)
	flasher := ui.NewFlasher(jsonFlag, stdoutFlag)

	writer, err := output.New(appContext)
	if err != nil {
		fmt.Printf("[!] Output Error: %v\\n", err)
		os.Exit(1)
	}
	defer writer.Close()

	jsScanner := jssecrethunter.New(jsScanFlag)
	if jsScanFlag && !jsScanner.IsAvailable() {
		fmt.Println("[!] Warning: JSSecretHunter not found. Run tools_install.sh to install.")
	}

	pipe := pipeline.New(appContext, deduplicator, checker, filterEngine, writer, stats, flasher.Notify, jsScanner)

	ctx := context.Background()
	fmt.Printf("[*] Target: %s\\n", appContext.Domain)
	if appContext.OutputDir != "" {
		fmt.Printf("[*] Output: %s\\n", appContext.OutputDir)
	}

	stopHUD := ui.StartHUD(ctx, stats, sourceMgr, jsonFlag, stdoutFlag)
	defer stopHUD()

	rawChan := sourceMgr.StartAll(ctx)
	done := pipe.Start(ctx, rawChan)

	<-done
	stats.PrintReport()
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(scanCmd)

	// REQUIRED
	rootCmd.PersistentFlags().StringVarP(&domainFlag, "domain", "d", "", "Target domain to scan (e.g., example.com)")
	rootCmd.PersistentFlags().StringVarP(&inputFlag, "input", "i", "", "Input file containing URLs or subdomains")
	rootCmd.PersistentFlags().StringVarP(&targetListFlag, "target-list", "t", "", "File containing list of target domains (one per line)")

	// ADVANCED
	rootCmd.PersistentFlags().BoolVar(&wildcardFlag, "wildcard", false, "Enable wildcard subdomain handling")
	rootCmd.PersistentFlags().BoolVar(&noDedupFlag, "no-dedup", false, "Disable deduplication")
	rootCmd.PersistentFlags().StringVar(&mcFlag, "mc", "", "Match Status Codes (e.g., 200,403,404)")
	rootCmd.PersistentFlags().StringVar(&sourcesFlag, "sources", "", "Comma-separated list of sources to use")
	rootCmd.PersistentFlags().BoolVar(&initConfigFlag, "init-config", false, "Create a default configuration file")

	// OUTPUT
	rootCmd.PersistentFlags().StringVarP(&outputFlag, "output", "o", "", "Output directory for results")
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output in JSON Lines format")
	rootCmd.PersistentFlags().BoolVar(&stdoutFlag, "stdout", false, "Stream output to stdout even if -o is set")

	// PERFORMANCE
	rootCmd.PersistentFlags().IntVarP(&workersFlag, "workers", "w", 20, "Number of concurrent workers")
	rootCmd.PersistentFlags().IntVar(&delayFlag, "delay", 0, "Delay between requests in milliseconds")
	rootCmd.PersistentFlags().IntVar(&timeoutFlag, "timeout", 10, "HTTP timeout in seconds")

	// FILTERS
	rootCmd.PersistentFlags().BoolVar(&sensitiveUrlsFlag, "sensitive-urls", false, "Filter for sensitive keywords (secrets, backups, etc.)")
	rootCmd.PersistentFlags().BoolVar(&paramsFlag, "params", false, "Filter for interesting parameters (SQLi, XSS, etc.)")
	rootCmd.PersistentFlags().BoolVar(&jsFlag, "js", false, "Filter for JavaScript files")
	rootCmd.PersistentFlags().BoolVar(&excludeLibsFlag, "exclude-libs", false, "Exclude common JS libraries (jquery, react, etc.)")
	rootCmd.PersistentFlags().BoolVar(&pdfFlag, "pdf", false, "Filter for PDF files")
	rootCmd.PersistentFlags().BoolVar(&logFlag, "log", false, "Filter for Log files")
	rootCmd.PersistentFlags().BoolVar(&configFilterFlag, "config", false, "Filter for Config files")

	// SCANNERS
	rootCmd.PersistentFlags().BoolVar(&jsScanFlag, "js-scan", false, "Run JSSecretHunter on discovered JS files")

	cobra.OnInitialize(config.Init)
}
