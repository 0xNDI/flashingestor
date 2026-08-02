package main

import (
	"context"
	"strings"
	"sync"

	"github.com/Macmod/flashingestor/config"
	"github.com/Macmod/flashingestor/core"
)

// runHeadless executes the requested collection/conversion steps without the
// TUI. Steps are always run in the canonical order ingest -> remote -> convert,
// independently of the order in cfg.Steps. After all steps finish the log
// channel is closed and the logger is allowed to drain so that all output is
// flushed before main returns.
func runHeadless(
	cfg *config.Config,
	logger *core.Logger,
	logChannel chan core.LogMessage,
	ingestMgr *IngestionManager,
	conversionMgr *ConversionManager,
	remoteMgr *RemoteCollectionManager,
	disableIngest bool,
	disableRemote bool,
	remoteNoCrossDomain bool,
	initialDomain string,
	initialBaseDN string,
	initialDC string,
) {
	enabled := parseSteps(cfg.Steps)
	if len(enabled) == 0 {
		enabled["ingest"] = true
		enabled["convert"] = true
	}

	logger.Log0("🖥️ [blue]Running in headless mode (no TUI). Steps:[-] %s", formatSteps(enabled))

	if enabled["ingest"] {
		if disableIngest {
			logger.Log0("🫠 [yellow]Skipping 'ingest' step (no ingestion credentials provided).[-]")
		} else if initialDomain == "" {
			logger.Log0("🫠 [yellow]Skipping 'ingest' step (could not determine initial domain).[-]")
		} else {
			// Start from a clean ingested-domains tracker and mark the initial
			// domain as processed (mirrors the interactive ingestion callback).
			ingestMgr.processedDomains = &sync.Map{}
			ingestMgr.processedDomains.Store(strings.ToUpper(initialDomain), true)

			ctx := context.Background()
			ingestMgr.start(ctx, initialDomain, initialBaseDN, initialDC)
			ingestMgr.Wait()
		}
	}

	if enabled["remote"] {
		if disableRemote {
			logger.Log0("🫠 [yellow]Skipping 'remote' step (no remote collection credentials provided).[-]")
		} else {
			remoteMgr.start(cfg.RemoteAuth, remoteNoCrossDomain)
			remoteMgr.Wait()
		}
	}

	if enabled["convert"] {
		conversionMgr.start()
		conversionMgr.Wait()
	}

	logger.Log0("✅ [green]Headless run finished.[-]")

	// Flush any remaining log messages before exiting.
	close(logChannel)
	<-logger.Done()
}

// parseSteps parses a comma-separated list of step names into a set.
// Keys are stored lowercase for case-insensitive matching.
func parseSteps(s string) map[string]bool {
	steps := make(map[string]bool)
	for _, name := range strings.Split(s, ",") {
		name = strings.ToLower(strings.TrimSpace(name))
		if name != "" {
			steps[name] = true
		}
	}
	return steps
}

// formatSteps renders the enabled steps as an ordered, human-readable string.
func formatSteps(enabled map[string]bool) string {
	var ordered []string
	for _, name := range []string{"ingest", "remote", "convert"} {
		if enabled[name] {
			ordered = append(ordered, name)
		}
	}
	return strings.Join(ordered, ", ")
}
