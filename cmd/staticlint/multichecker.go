// Package main defines a custom multichecker tool for static analysis of Go code.
//
// This tool combines a number of analyzers from the Go standard toolset,
// Staticcheck (honnef.co/go/tools), and a custom analyzer `noosexit` that
// prevents direct calls to os.Exit in the main function of the main package.
//
// Usage:
//
//  1. Build the binary:
//     go build -o bin/staticlint ./cmd/staticlint
//
//  2. Run the tool on a Go package:
//     ./bin/staticlint ./...
//
// The multichecker is initialized using `analysis/multichecker.Main`
// which accepts a variadic list of analyzers. These analyzers are executed
// sequentially over the input packages.
package main

import (
	"github.com/vadicheck/shorturl/internal/analyzer/noosexit"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
)

// getAnalyzers returns all analyzers used in the multichecker — for testability
func getAnalyzers() []*analysis.Analyzer {
	// A map of selected Staticcheck analyzers to include.
	// Only these analyzers will be added from the Staticcheck set.
	checks := map[string]bool{
		"SA1012": true, // nil context passed to functions
		"SA4006": true, // unused value assigned to a variable
		"SA5000": true, // unreachable code
		"SA6000": true, // using regexp.Match instead of compiling
		"SA9004": true, // unnecessary conversion
	}

	var allChecks []*analysis.Analyzer

	// Add selected Staticcheck analyzers from honnef.co/go/tools.
	for _, v := range staticcheck.Analyzers {
		if checks[v.Analyzer.Name] {
			allChecks = append(allChecks, v.Analyzer)
		}
	}

	// Add standard Go analyzers.
	allChecks = append(allChecks,
		printf.Analyzer,       // checks formatting directives in Printf-like functions
		shadow.Analyzer,       // detects shadowed variables
		structtag.Analyzer,    // detects incorrect struct tags
		unreachable.Analyzer,  // detects unreachable code
		unusedresult.Analyzer, // detects results of calls that are unused
		noosexit.Analyzer,
	)

	return allChecks
}

func main() {
	// Run all analyzers via multichecker.
	multichecker.Main(getAnalyzers()...)
}
