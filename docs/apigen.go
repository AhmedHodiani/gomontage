package docs

import (
	"fmt"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"sort"
	"strings"
)

// packageDoc holds the parsed documentation for a single Go package.
type packageDoc struct {
	Name       string
	ImportPath string
	Doc        string
	Types      []*typeDoc
	Funcs      []*funcDoc
	Constants  []*constDoc
	Vars       []*varDoc
}

// typeDoc holds documentation for a single exported type.
type typeDoc struct {
	Name         string
	Doc          string
	Signature    string
	Constructors []*funcDoc
	Methods      []*funcDoc
	Fields       []*fieldDoc
	IsInterface  bool
	IsStruct     bool
}

// funcDoc holds documentation for a single exported function or method.
type funcDoc struct {
	Name      string
	Doc       string
	Signature string
	Receiver  string
}

// fieldDoc holds documentation for a single struct/interface field.
type fieldDoc struct {
	Name string
	Type string
	Doc  string
}

// constDoc holds documentation for a constant or constant group.
type constDoc struct {
	Name  string
	Type  string
	Value string
	Doc   string
}

// varDoc holds documentation for a variable.
type varDoc struct {
	Name string
	Type string
	Doc  string
}

// parsePackage parses a Go package directory and returns structured documentation.
func parsePackage(dir, importPath string) (*packageDoc, error) {
	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		// Skip test files.
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", dir, err)
	}

	// Take the first (usually only) package.
	var astPkg *ast.Package
	for _, p := range pkgs {
		astPkg = p
		break
	}
	if astPkg == nil {
		return nil, fmt.Errorf("no Go package found in %s", dir)
	}

	dpkg := doc.New(astPkg, importPath, doc.AllDecls)

	result := &packageDoc{
		Name:       dpkg.Name,
		ImportPath: importPath,
		Doc:        dpkg.Doc,
	}

	// Parse types.
	for _, dt := range dpkg.Types {
		if !token.IsExported(dt.Name) {
			continue
		}
		td := &typeDoc{
			Name: dt.Name,
			Doc:  dt.Doc,
		}

		// Extract type signature and fields.
		for _, spec := range dt.Decl.Specs {
			if ts, ok := spec.(*ast.TypeSpec); ok {
				td.Signature = formatNode(fset, ts)
				td.IsInterface = isInterface(ts)
				td.IsStruct = isStruct(ts)
				td.Fields = extractFields(fset, ts)
			}
		}

		// Constructors (associated functions that return this type).
		for _, fn := range dt.Funcs {
			if !token.IsExported(fn.Name) {
				continue
			}
			td.Constructors = append(td.Constructors, &funcDoc{
				Name:      fn.Name,
				Doc:       fn.Doc,
				Signature: formatFuncDecl(fset, fn.Decl),
			})
		}

		// Methods.
		for _, fn := range dt.Methods {
			if !token.IsExported(fn.Name) {
				continue
			}
			td.Methods = append(td.Methods, &funcDoc{
				Name:      fn.Name,
				Doc:       fn.Doc,
				Signature: formatFuncDecl(fset, fn.Decl),
				Receiver:  dt.Name,
			})
		}

		result.Types = append(result.Types, td)
	}

	// Parse package-level functions (not associated with a type).
	for _, fn := range dpkg.Funcs {
		if !token.IsExported(fn.Name) {
			continue
		}
		result.Funcs = append(result.Funcs, &funcDoc{
			Name:      fn.Name,
			Doc:       fn.Doc,
			Signature: formatFuncDecl(fset, fn.Decl),
		})
	}

	// Parse constants.
	for _, c := range dpkg.Consts {
		for _, spec := range c.Decl.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range vs.Names {
				if !token.IsExported(name.Name) {
					continue
				}
				cd := &constDoc{
					Name: name.Name,
					Doc:  vs.Doc.Text(),
				}
				if vs.Type != nil {
					cd.Type = formatNode(fset, vs.Type)
				}
				if i < len(vs.Values) {
					cd.Value = formatNode(fset, vs.Values[i])
				}
				result.Constants = append(result.Constants, cd)
			}
		}
	}

	// Parse variables.
	for _, v := range dpkg.Vars {
		for _, spec := range v.Decl.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range vs.Names {
				if !token.IsExported(name.Name) {
					continue
				}
				vd := &varDoc{
					Name: name.Name,
					Doc:  vs.Doc.Text(),
				}
				if vs.Type != nil {
					vd.Type = formatNode(fset, vs.Type)
				}
				result.Vars = append(result.Vars, vd)
			}
		}
	}

	return result, nil
}

// formatNode formats an AST node back to Go source code.
func formatNode(fset *token.FileSet, node ast.Node) string {
	var buf strings.Builder
	printer.Fprint(&buf, fset, node)
	return buf.String()
}

// formatFuncDecl formats a function declaration as a clean signature string.
func formatFuncDecl(fset *token.FileSet, decl *ast.FuncDecl) string {
	// Build the signature manually for cleaner output.
	var buf strings.Builder
	buf.WriteString("func ")

	// Receiver.
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		buf.WriteString("(")
		buf.WriteString(formatNode(fset, decl.Recv.List[0].Type))
		buf.WriteString(") ")
	}

	buf.WriteString(decl.Name.Name)

	// Parameters.
	buf.WriteString("(")
	buf.WriteString(formatFieldList(fset, decl.Type.Params))
	buf.WriteString(")")

	// Return values.
	if decl.Type.Results != nil && len(decl.Type.Results.List) > 0 {
		results := formatFieldList(fset, decl.Type.Results)
		if len(decl.Type.Results.List) == 1 && len(decl.Type.Results.List[0].Names) == 0 {
			buf.WriteString(" ")
			buf.WriteString(results)
		} else {
			buf.WriteString(" (")
			buf.WriteString(results)
			buf.WriteString(")")
		}
	}

	return buf.String()
}

// formatFieldList formats a field list (parameters or results) as a Go-style string.
func formatFieldList(fset *token.FileSet, fl *ast.FieldList) string {
	if fl == nil {
		return ""
	}
	var parts []string
	for _, field := range fl.List {
		typeStr := formatNode(fset, field.Type)
		if len(field.Names) == 0 {
			parts = append(parts, typeStr)
		} else {
			names := make([]string, len(field.Names))
			for i, n := range field.Names {
				names[i] = n.Name
			}
			parts = append(parts, strings.Join(names, ", ")+" "+typeStr)
		}
	}
	return strings.Join(parts, ", ")
}

// isInterface checks if a TypeSpec defines an interface.
func isInterface(ts *ast.TypeSpec) bool {
	_, ok := ts.Type.(*ast.InterfaceType)
	return ok
}

// isStruct checks if a TypeSpec defines a struct.
func isStruct(ts *ast.TypeSpec) bool {
	_, ok := ts.Type.(*ast.StructType)
	return ok
}

// extractFields extracts exported fields from a struct or interface type.
func extractFields(fset *token.FileSet, ts *ast.TypeSpec) []*fieldDoc {
	var fields []*fieldDoc

	switch t := ts.Type.(type) {
	case *ast.StructType:
		if t.Fields == nil {
			break
		}
		for _, field := range t.Fields.List {
			// Skip embedded fields (they have no names).
			if len(field.Names) == 0 {
				continue
			}
			for _, name := range field.Names {
				if !token.IsExported(name.Name) {
					continue
				}
				fields = append(fields, &fieldDoc{
					Name: name.Name,
					Type: formatNode(fset, field.Type),
					Doc:  field.Doc.Text(),
				})
			}
		}
	case *ast.InterfaceType:
		if t.Methods == nil {
			break
		}
		for _, method := range t.Methods.List {
			if len(method.Names) == 0 {
				continue
			}
			for _, name := range method.Names {
				if !token.IsExported(name.Name) {
					continue
				}
				fields = append(fields, &fieldDoc{
					Name: name.Name,
					Type: formatNode(fset, method.Type),
					Doc:  method.Doc.Text(),
				})
			}
		}
	}

	return fields
}

// generateAPIPage renders a packageDoc into a markdown string for the API reference.
func generateAPIPage(pkg *packageDoc) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("# Package %s\n\n", pkg.Name))
	b.WriteString(fmt.Sprintf("    import \"%s\"\n\n", pkg.ImportPath))

	// Package overview.
	if pkg.Doc != "" {
		b.WriteString(formatDocComment(pkg.Doc))
		b.WriteString("\n")
	}

	// Separate interfaces, structs, and other types.
	var interfaces []*typeDoc
	var structs []*typeDoc
	for _, t := range pkg.Types {
		if t.IsInterface {
			interfaces = append(interfaces, t)
		} else {
			structs = append(structs, t)
		}
	}

	// Constants.
	if len(pkg.Constants) > 0 {
		b.WriteString("## Constants\n\n")
		// Group by type if possible.
		grouped := groupConstsByType(pkg.Constants)
		for _, group := range grouped {
			if group.typeName != "" {
				b.WriteString(fmt.Sprintf("### %s\n\n", group.typeName))
			}
			b.WriteString("| Name | Value | Description |\n")
			b.WriteString("|------|-------|-------------|\n")
			for _, c := range group.consts {
				doc := firstSentence(c.Doc)
				value := c.Value
				if value == "" {
					value = "-"
				}
				b.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", c.Name, value, doc))
			}
			b.WriteString("\n")
		}
	}

	// Variables.
	if len(pkg.Vars) > 0 {
		b.WriteString("## Variables\n\n")
		b.WriteString("| Name | Description |\n")
		b.WriteString("|------|-------------|\n")
		for _, v := range pkg.Vars {
			doc := firstSentence(v.Doc)
			b.WriteString(fmt.Sprintf("| `%s` | %s |\n", v.Name, doc))
		}
		b.WriteString("\n")
	}

	// Package-level functions.
	if len(pkg.Funcs) > 0 {
		b.WriteString("## Functions\n\n")
		for _, fn := range pkg.Funcs {
			writeFunc(&b, fn)
		}
	}

	// Interfaces.
	if len(interfaces) > 0 {
		b.WriteString("## Interfaces\n\n")
		for _, iface := range interfaces {
			writeTypeDoc(&b, iface)
		}
	}

	// Types (structs).
	if len(structs) > 0 {
		b.WriteString("## Types\n\n")
		for _, s := range structs {
			writeTypeDoc(&b, s)
		}
	}

	return b.String()
}

// writeTypeDoc writes the documentation for a single type.
func writeTypeDoc(b *strings.Builder, td *typeDoc) {
	b.WriteString(fmt.Sprintf("### %s\n\n", td.Name))

	if td.Doc != "" {
		b.WriteString(formatDocComment(td.Doc))
		b.WriteString("\n")
	}

	// Show type definition.
	if td.Signature != "" {
		b.WriteString("```go\n")
		b.WriteString(fmt.Sprintf("type %s\n", td.Signature))
		b.WriteString("```\n\n")
	}

	// Show exported fields for structs.
	if len(td.Fields) > 0 && !td.IsInterface {
		b.WriteString("#### Fields\n\n")
		b.WriteString("| Field | Type | Description |\n")
		b.WriteString("|-------|------|-------------|\n")
		for _, f := range td.Fields {
			doc := firstSentence(f.Doc)
			b.WriteString(fmt.Sprintf("| `%s` | `%s` | %s |\n", f.Name, f.Type, doc))
		}
		b.WriteString("\n")
	}

	// Show interface methods.
	if len(td.Fields) > 0 && td.IsInterface {
		b.WriteString("#### Methods\n\n")
		b.WriteString("| Method | Signature | Description |\n")
		b.WriteString("|--------|-----------|-------------|\n")
		for _, f := range td.Fields {
			doc := firstSentence(f.Doc)
			sig := formatInterfaceMethodSig(f.Name, f.Type)
			b.WriteString(fmt.Sprintf("| `%s` | `%s` | %s |\n", f.Name, sig, doc))
		}
		b.WriteString("\n")
	}

	// Constructors.
	if len(td.Constructors) > 0 {
		b.WriteString("#### Constructors\n\n")
		for _, fn := range td.Constructors {
			writeFunc(b, fn)
		}
	}

	// Methods (for structs / concrete types).
	if len(td.Methods) > 0 {
		b.WriteString("#### Methods\n\n")
		b.WriteString("| Method | Signature | Description |\n")
		b.WriteString("|--------|-----------|-------------|\n")
		for _, fn := range td.Methods {
			doc := firstSentence(fn.Doc)
			b.WriteString(fmt.Sprintf("| `%s` | `%s` | %s |\n", fn.Name, fn.Signature, doc))
		}
		b.WriteString("\n")
		// Write detailed docs per method.
		for _, fn := range td.Methods {
			if hasDetailedDoc(fn.Doc) {
				b.WriteString(fmt.Sprintf("##### %s\n\n", fn.Name))
				b.WriteString(fmt.Sprintf("```go\n%s\n```\n\n", fn.Signature))
				b.WriteString(formatDocComment(fn.Doc))
				b.WriteString("\n")
			}
		}
	}
}

// writeFunc writes documentation for a single function.
func writeFunc(b *strings.Builder, fn *funcDoc) {
	b.WriteString(fmt.Sprintf("#### %s\n\n", fn.Name))
	b.WriteString(fmt.Sprintf("```go\n%s\n```\n\n", fn.Signature))
	if fn.Doc != "" {
		b.WriteString(formatDocComment(fn.Doc))
		b.WriteString("\n")
	}
}

// constGroup groups constants by their declared type.
type constGroup struct {
	typeName string
	consts   []*constDoc
}

// groupConstsByType groups constants by their type name.
func groupConstsByType(consts []*constDoc) []constGroup {
	m := make(map[string][]*constDoc)
	var order []string
	for _, c := range consts {
		key := c.Type
		if _, exists := m[key]; !exists {
			order = append(order, key)
		}
		m[key] = append(m[key], c)
	}

	var groups []constGroup
	for _, key := range order {
		groups = append(groups, constGroup{typeName: key, consts: m[key]})
	}
	return groups
}

// formatDocComment formats a Go doc comment for markdown output.
// It preserves code examples (indented blocks) and wraps them in fences.
func formatDocComment(doc string) string {
	if doc == "" {
		return ""
	}

	lines := strings.Split(strings.TrimRight(doc, "\n"), "\n")
	var result strings.Builder
	inCode := false

	for _, line := range lines {
		// Go doc comments use a tab indent for code examples.
		if strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "    ") {
			if !inCode {
				result.WriteString("```go\n")
				inCode = true
			}
			// Remove one level of indentation.
			trimmed := strings.TrimPrefix(line, "\t")
			trimmed = strings.TrimPrefix(trimmed, "    ")
			result.WriteString(trimmed)
			result.WriteString("\n")
		} else {
			if inCode {
				result.WriteString("```\n\n")
				inCode = false
			}
			result.WriteString(line)
			result.WriteString("\n")
		}
	}
	if inCode {
		result.WriteString("```\n")
	}

	return result.String()
}

// formatInterfaceMethodSig formats an interface method signature for display.
// The AST produces f.Name = "MethodName" and f.Type = "func() ReturnType",
// so the proper combined signature is "MethodName() ReturnType".
func formatInterfaceMethodSig(name, funcType string) string {
	// f.Type from the AST looks like "func(params) returns".
	// Strip the leading "func" to get "(params) returns", then prepend name.
	if strings.HasPrefix(funcType, "func") {
		return name + funcType[len("func"):]
	}
	// Fallback: just separate with a space.
	return name + " " + funcType
}

// firstSentence extracts the first sentence from a doc comment.
func firstSentence(doc string) string {
	if doc == "" {
		return ""
	}
	doc = strings.TrimSpace(doc)

	// Take the first line.
	if idx := strings.Index(doc, "\n"); idx >= 0 {
		doc = doc[:idx]
	}

	doc = strings.TrimSpace(doc)
	// Remove trailing period if present (for cleaner table display).
	doc = strings.TrimSuffix(doc, ".")
	return doc
}

// hasDetailedDoc checks if a doc comment has more than a single-line summary.
func hasDetailedDoc(doc string) bool {
	if doc == "" {
		return false
	}
	lines := strings.Split(strings.TrimSpace(doc), "\n")
	// Consider it detailed if it has more than 2 lines or contains a code example.
	if len(lines) > 2 {
		return true
	}
	return strings.Contains(doc, "\t") // Tab-indented code block.
}

// generateAPIIndex generates the API index page that links to all package pages.
func generateAPIIndex(packages []*packageDoc) string {
	var b strings.Builder

	b.WriteString("# API Reference\n\n")
	b.WriteString("Complete API documentation auto-generated from source code.\n\n")

	b.WriteString("## Packages\n\n")
	b.WriteString("| Package | Description |\n")
	b.WriteString("|---------|-------------|\n")

	for _, pkg := range packages {
		desc := firstSentence(pkg.Doc)
		filename := apiFilename(pkg.Name)
		b.WriteString(fmt.Sprintf("| [`%s`](%s) | %s |\n", pkg.Name, filename, desc))
	}

	b.WriteString("\n")

	// Summary of all types across packages.
	b.WriteString("## Quick Reference\n\n")
	for _, pkg := range packages {
		if len(pkg.Types) == 0 && len(pkg.Funcs) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("### %s\n\n", pkg.Name))

		if len(pkg.Types) > 0 {
			// Sort: interfaces first, then structs.
			sorted := make([]*typeDoc, len(pkg.Types))
			copy(sorted, pkg.Types)
			sort.Slice(sorted, func(i, j int) bool {
				if sorted[i].IsInterface != sorted[j].IsInterface {
					return sorted[i].IsInterface
				}
				return sorted[i].Name < sorted[j].Name
			})

			b.WriteString("| Type | Kind | Description |\n")
			b.WriteString("|------|------|-------------|\n")
			for _, t := range sorted {
				kind := "type"
				if t.IsInterface {
					kind = "interface"
				} else if t.IsStruct {
					kind = "struct"
				}
				desc := firstSentence(t.Doc)
				b.WriteString(fmt.Sprintf("| `%s` | %s | %s |\n", t.Name, kind, desc))
			}
			b.WriteString("\n")
		}

		if len(pkg.Funcs) > 0 {
			b.WriteString("| Function | Description |\n")
			b.WriteString("|----------|-------------|\n")
			for _, fn := range pkg.Funcs {
				desc := firstSentence(fn.Doc)
				b.WriteString(fmt.Sprintf("| `%s` | %s |\n", fn.Name, desc))
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}

// apiFilename returns the filename for a package's API reference page.
func apiFilename(pkgName string) string {
	return fmt.Sprintf("api-%s.md", pkgName)
}
