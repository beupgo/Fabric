// voltagent - A utility for validating and checking Fabric configurations
//
// Usage:
//   voltagent [options]
//
// Examples:
//   voltagent                                  # Check current Fabric configuration
//   voltagent --version                        # Show version
//   voltagent --check-patterns                 # Validate pattern files
//   voltagent --config /path/to/fabric         # Check specific config directory
//
// Description:
//   voltagent is a helper tool for the Fabric AI framework that validates
//   configurations, checks pattern files, and reports potential issues.

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var version = "v1.0.0"

func main() {
	// Define command-line flags
	showVersion := flag.Bool("version", false, "Show version information")
	checkPatterns := flag.Bool("check-patterns", false, "Validate pattern files")
	configDir := flag.String("config", "", "Fabric config directory (default: ~/.config/fabric)")
	flag.Usage = printUsage
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("voltagent version %s\n", version)
		os.Exit(0)
	}

	// Determine config directory
	fabricConfig := *configDir
	if fabricConfig == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Cannot determine home directory: %v\n", err)
			os.Exit(1)
		}
		fabricConfig = filepath.Join(home, ".config", "fabric")
	}

	// Handle check-patterns flag
	if *checkPatterns {
		checkPatternFiles(fabricConfig)
		return
	}

	// Default behavior: check basic configuration
	checkConfiguration(fabricConfig)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `voltagent - Fabric configuration validator

Usage:
  voltagent [options]

Options:
`)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
Examples:
  voltagent                     # Check Fabric configuration
  voltagent --version           # Show version
  voltagent --check-patterns    # Validate pattern files
`)
}

func checkConfiguration(configDir string) {
	fmt.Printf("Checking Fabric configuration in: %s\n\n", configDir)

	// Check if config directory exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "❌ Configuration directory not found: %s\n", configDir)
		fmt.Fprintf(os.Stderr, "   Run 'fabric --setup' to initialize Fabric\n")
		os.Exit(1)
	}
	fmt.Printf("✓ Configuration directory exists\n")

	// Check for patterns directory
	patternsDir := filepath.Join(configDir, "patterns")
	if info, err := os.Stat(patternsDir); err == nil && info.IsDir() {
		// Count patterns
		count := countPatterns(patternsDir)
		fmt.Printf("✓ Patterns directory exists (%d patterns found)\n", count)
	} else {
		fmt.Printf("⚠ Patterns directory not found\n")
	}

	// Check for .env file
	envFile := filepath.Join(configDir, ".env")
	if _, err := os.Stat(envFile); err == nil {
		fmt.Printf("✓ Environment file exists\n")
	} else {
		fmt.Printf("⚠ Environment file not found (.env)\n")
	}

	fmt.Printf("\n✅ Configuration check complete\n")
}

func countPatterns(patternsDir string) int {
	count := 0
	entries, err := os.ReadDir(patternsDir)
	if err != nil {
		return 0
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// Check if directory contains system.md
			systemFile := filepath.Join(patternsDir, entry.Name(), "system.md")
			if _, err := os.Stat(systemFile); err == nil {
				count++
			}
		}
	}
	return count
}

func checkPatternFiles(configDir string) {
	patternsDir := filepath.Join(configDir, "patterns")

	fmt.Printf("Validating pattern files in: %s\n\n", patternsDir)

	if _, err := os.Stat(patternsDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "❌ Patterns directory not found: %s\n", patternsDir)
		os.Exit(1)
	}

	entries, err := os.ReadDir(patternsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading patterns directory: %v\n", err)
		os.Exit(1)
	}

	validPatterns := 0
	invalidPatterns := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		patternName := entry.Name()
		patternDir := filepath.Join(patternsDir, patternName)
		systemFile := filepath.Join(patternDir, "system.md")

		// Check if system.md exists
		if _, err := os.Stat(systemFile); os.IsNotExist(err) {
			fmt.Printf("❌ %s: Missing system.md\n", patternName)
			invalidPatterns++
			continue
		}

		// Check if system.md is not empty
		content, err := os.ReadFile(systemFile)
		if err != nil {
			fmt.Printf("❌ %s: Cannot read system.md: %v\n", patternName, err)
			invalidPatterns++
			continue
		}

		if len(strings.TrimSpace(string(content))) == 0 {
			fmt.Printf("❌ %s: system.md is empty\n", patternName)
			invalidPatterns++
			continue
		}

		validPatterns++
	}

	fmt.Printf("\n📊 Results:\n")
	fmt.Printf("   Valid patterns: %d\n", validPatterns)
	if invalidPatterns > 0 {
		fmt.Printf("   Invalid patterns: %d\n", invalidPatterns)
	}

	if invalidPatterns > 0 {
		os.Exit(1)
	}

	fmt.Printf("\n✅ All patterns validated successfully\n")
}
