// Package docs provides automated API documentation generation for Gomontage.
//
// The generator produces markdown files covering all public APIs of the
// framework. Guide pages (getting started, clips, timeline, effects, export)
// are maintained as hand-written markdown templates. API reference pages
// are auto-generated from Go source code using AST parsing, extracting
// every exported type, function, method, and constant along with their
// doc comments and full signatures.
//
// Usage:
//
//	gomontage docs              # Generates docs/ directory
//	gomontage docs -o api-docs  # Custom output directory
package docs

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed templates/*.md
var templates embed.FS

// apiPackage describes a package to generate API docs for.
type apiPackage struct {
	// dir is the directory path relative to the module root.
	dir string
	// importPath is the full Go import path.
	importPath string
}

// packages defines which packages get API reference docs generated.
// Engine is excluded — it is an internal layer not intended for direct user use.
var packages = []apiPackage{
	{".", "github.com/ahmedhodiani/gomontage"},
	{"clip", "github.com/ahmedhodiani/gomontage/clip"},
	{"timeline", "github.com/ahmedhodiani/gomontage/timeline"},
	{"effects", "github.com/ahmedhodiani/gomontage/effects"},
	{"export", "github.com/ahmedhodiani/gomontage/export"},
}

// Generate produces markdown documentation files in the specified output directory.
// It copies hand-written guide templates and auto-generates API reference pages
// from the Go source code.
func Generate(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("could not create docs directory: %w", err)
	}

	// Phase 1: Copy guide templates to the output directory.
	if err := copyTemplates(outputDir); err != nil {
		return err
	}

	// Phase 2: Auto-generate API reference pages from source.
	moduleRoot, err := findModuleRoot()
	if err != nil {
		return fmt.Errorf("could not find module root: %w", err)
	}

	var allPkgs []*packageDoc
	for _, pkg := range packages {
		dir := filepath.Join(moduleRoot, pkg.dir)
		pdoc, err := parsePackage(dir, pkg.importPath)
		if err != nil {
			return fmt.Errorf("parsing package %s: %w", pkg.importPath, err)
		}
		allPkgs = append(allPkgs, pdoc)

		// Write per-package API page.
		filename := apiFilename(pdoc.Name)
		content := generateAPIPage(pdoc)
		path := filepath.Join(outputDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("could not write %s: %w", filename, err)
		}
		fmt.Printf("  Generated %s\n", filename)
	}

	// Phase 3: Generate the API index page.
	indexContent := generateAPIIndex(allPkgs)
	indexPath := filepath.Join(outputDir, "api-reference.md")
	if err := os.WriteFile(indexPath, []byte(indexContent), 0644); err != nil {
		return fmt.Errorf("could not write api-reference.md: %w", err)
	}
	fmt.Printf("  Generated api-reference.md\n")

	return nil
}

// copyTemplates copies embedded guide markdown files to the output directory.
func copyTemplates(outputDir string) error {
	entries, err := templates.ReadDir("templates")
	if err != nil {
		return fmt.Errorf("reading embedded templates: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := templates.ReadFile("templates/" + entry.Name())
		if err != nil {
			return fmt.Errorf("reading template %s: %w", entry.Name(), err)
		}

		outPath := filepath.Join(outputDir, entry.Name())
		if err := os.WriteFile(outPath, data, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", entry.Name(), err)
		}
		fmt.Printf("  Copied %s\n", entry.Name())
	}

	return nil
}

// findModuleRoot locates the Gomontage module root directory.
// It uses runtime.Caller to find the docs package source location
// and walks up to the module root (directory containing go.mod).
func findModuleRoot() (string, error) {
	// Use this file's location as a starting point.
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("could not determine source file location")
	}

	// This file is at <module_root>/docs/generator.go, so the module root
	// is one directory up.
	dir := filepath.Dir(thisFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fallback: try relative to the working directory.
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("could not determine working directory: %w", err)
	}

	// Check common locations.
	for _, candidate := range []string{wd, filepath.Join(wd, "..")} {
		if _, err := os.Stat(filepath.Join(candidate, "go.mod")); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("could not find go.mod in any parent directory of %s", thisFile)
}

// guideFiles returns the guide template filenames for test assertions.
func guideFiles() []string {
	entries, err := templates.ReadDir("templates")
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			names = append(names, e.Name())
		}
	}
	return names
}
