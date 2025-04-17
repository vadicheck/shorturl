// Package noosexit defines a custom analyzer that checks for the use of
// os.Exit in the main.main function.
//
// This analyzer ensures that the main function of the `main` package does
// not call `os.Exit`, as it's considered bad practice for graceful application
// shutdown and signal handling in a Go program.
//
// Usage:
//
// The analyzer is intended to be added to a multichecker tool, which runs
// over a set of Go packages and performs various static analyses. This analyzer
// reports an error whenever `os.Exit` is used in the `main` function of the `main`
// package. Instead of calling `os.Exit`, it is better to handle errors and
// exit gracefully, e.g., using context cancellation or deferred cleanup.
package noosexit

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Analyzer is the main entry point for this custom analysis.
// It checks that os.Exit is not used in main.main().
var Analyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "checks that os.Exit is not used in main.main()",
	Run:  run,
}

// run is the function that performs the actual analysis of the Go source files.
// It inspects each file in the package and checks if the main function calls
// os.Exit. If found, it reports an error.
//
// It skips files under "go-build" directory and files that are not in the `main` package.
// If `os.Exit` is found within the `main.main` function, a report is generated.
func run(pass *analysis.Pass) (interface{}, error) {
	// Iterate over all the files in the package
	for _, file := range pass.Files {
		// Skip files under the "go-build" directory
		if fullPath := pass.Fset.Position(file.Pos()).String(); strings.Contains(fullPath, "go-build") {
			continue
		}

		// Ensure we're checking the main package
		if pass.Pkg.Name() != "main" {
			continue
		}

		// Inspect each function in the file
		ast.Inspect(file, func(n ast.Node) bool {
			// Check if the function is "main"
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" {
				return true
			}

			// Inspect the body of the main function
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				// Look for function calls
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				// Check if the call is to os.Exit
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := sel.X.(*ast.Ident); ok &&
						ident.Name == "os" &&
						sel.Sel.Name == "Exit" {

						// Ensure that the package is the "os" package
						if obj := pass.TypesInfo.Uses[ident]; obj != nil {
							if pkg, ok := obj.(*types.PkgName); ok && pkg.Imported().Path() == "os" {
								// Report an error if os.Exit is found
								pass.Reportf(call.Pos(), "use of os.Exit in main.main is forbidden")
							}
						}
					}
				}
				return true
			})
			return false
		})
	}
	return nil, nil
}
