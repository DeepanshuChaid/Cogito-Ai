package commands

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileMap struct {
	Path           string   `json:"path,omitempty"`
	Importance     int      `json:"importance,omitempty"`
	Role           string   `json:"role,omitempty"`
	Summary        string   `json:"summary,omitempty"`
	KeyFunctions   []string `json:"key_functions,omitempty"`
	ImportantCalls []string `json:"important_calls,omitempty"`

	// Internal processing fields
	Package    string         `json:"-"`
	Imports    []string       `json:"-"`
	Functions  []Function     `json:"-"`
	Structs    []string       `json:"-"`
	Interfaces []string       `json:"-"`
	Methods    []Method       `json:"-"`
	Language   string         `json:"-"`
	Classes    []string       `json:"-"`
	Calls      []CallRelation `json:"-"`
	Ignore     bool           `json:"-"`
}

type CallRelation struct {
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`
}


type Function struct {
	Name string `json:"name,omitempty"`
	Line int   `json:"line,omitempty"`
}

type Method struct {
	Receiver string `json:"receiver,omitempty"`
	Name     string `json:"name,omitempty"`
}

type CodebaseMap struct {
	Entrypoints []string          `json:"entrypoints,omitempty"`
	StateMap    map[string]string `json:"state_map,omitempty"`
	Files       []FileMap         `json:"files,omitempty"`
	Flows       []string          `json:"flows,omitempty"`
}

func BuildMap() {
	root := "."

	var result CodebaseMap

	fmt.Println("Building codebase map...")

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}

		if info.IsDir() {
			name := info.Name()

			if ShouldSkipDir(name) {
				return filepath.SkipDir
			}

			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))

		supported := map[string]bool{
			".go":   true,
			".py":   true,
			".js":   true,
			".jsx":  true,
			".ts":   true,
			".tsx":  true,
			".java": true,
		}

		if !supported[ext] {
			return nil
		}

		fileMap := parseFile(path)
		fileMap.Role = classifyRole(path, &fileMap)
		fileMap.Ignore = isIgnoreZone(path)

		result.Files = append(result.Files, fileMap)

		return nil
	})

	// Pass 2: Calculate global metrics and high-signal structures
	inboundCalls := make(map[string]int)
	funcToFile := make(map[string]string)
	stateMap := make(map[string]string)
	entrypoints := []string{}

	for _, f := range result.Files {
		if isEntryPoint(f.Path, &f) {
			entrypoints = append(entrypoints, f.Path)
		}
		for _, fn := range f.Functions {
			funcToFile[fn.Name] = f.Path
		}
		// Detect state ownership
		role := classifyRole(f.Path, &f)
		if role == "state-manager" || role == "persistence-layer" || role == "config-layer" {
			key := "data"
			if strings.Contains(f.Path, "session") {
				key = "sessions"
			} else if strings.Contains(f.Path, "observation") || strings.Contains(f.Path, "memory") {
				key = "memory"
			} else if strings.Contains(f.Path, "config") {
				key = "config"
			}
			stateMap[key] = f.Path
		}
	}

	for _, f := range result.Files {
		for _, call := range f.Calls {
			if targetPath, ok := funcToFile[call.To]; ok {
				if targetPath != f.Path {
					inboundCalls[targetPath]++
				}
			}
		}
	}

	// Update importance and roles
	for i := range result.Files {
		f := &result.Files[i]
		f.Role = classifyRole(f.Path, f) // Re-classify with full context
		f.Importance = calculateImportance(f, inboundCalls[f.Path])

		// Signal logic
		for _, fn := range f.Functions {
			if isHighSignalFunc(fn.Name) {
				f.KeyFunctions = append(f.KeyFunctions, fn.Name)
			}
		}
		uniqueCalls := make(map[string]bool)
		for _, call := range f.Calls {
			if _, ok := funcToFile[call.To]; ok {
				if !uniqueCalls[call.To] {
					f.ImportantCalls = append(f.ImportantCalls, call.To)
					uniqueCalls[call.To] = true
				}
			}
		}
	}

	result.Entrypoints = entrypoints
	result.StateMap = stateMap

	// Sort files by importance (descending)
	sort.Slice(result.Files, func(i, j int) bool {
		if result.Files[i].Importance != result.Files[j].Importance {
			return result.Files[i].Importance > result.Files[j].Importance
		}
		return result.Files[i].Path < result.Files[j].Path
	})

	// Generate execution flows
	result.Flows = generateExecutionFlows(result.Files)

	os.MkdirAll(".cogito", os.ModePerm)

	fmt.Println("Creating map file...")

	file, err := os.Create(".cogito/map.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating map file: %v\n", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(result)
}

func parseGoFile(path string) FileMap {
	fset := token.NewFileSet()

	fmt.Println("Parsing file:", path)

	node, err := parser.ParseFile(
		fset,
		path,
		nil,
		parser.ParseComments,
	)

	if err != nil {
		return FileMap{
			Path: path,
		}
	}

	fileMap := FileMap{
		Path:    path,
		Package: node.Name.Name,
	}

	if node.Doc != nil {
		fileMap.Summary = strings.TrimSpace(node.Doc.Text())
	} else {
		// Fallback to searching first comment in file
		for _, cg := range node.Comments {
			if cg.Pos() < node.Name.Pos() {
				fileMap.Summary = strings.TrimSpace(cg.Text())
				break
			}
		}
	}

	for _, imp := range node.Imports {
		fileMap.Imports = append(
			fileMap.Imports,
			strings.Trim(imp.Path.Value, `"`),
		)
	}

	for _, decl := range node.Decls {
		switch d := decl.(type) {

		case *ast.FuncDecl:
			funcName := d.Name.Name
			if d.Recv == nil {
				fileMap.Functions = append(
					fileMap.Functions,
					Function{
						Name: funcName,
						Line: fset.Position(d.Pos()).Line,
					},
				)
			} else {
				receiver := ""
				if len(d.Recv.List) > 0 {
					switch r := d.Recv.List[0].Type.(type) {
					case *ast.Ident:
						receiver = r.Name
					case *ast.StarExpr:
						if ident, ok := r.X.(*ast.Ident); ok {
							receiver = ident.Name
						}
					}
				}
				fileMap.Methods = append(
					fileMap.Methods,
					Method{
						Receiver: receiver,
						Name:     funcName,
					},
				)
			}

			// Detect calls within the function body
			if d.Body != nil {
				ast.Inspect(d.Body, func(n ast.Node) bool {
					call, ok := n.(*ast.CallExpr)
					if !ok {
						return true
					}

					var calleeName string
					switch fun := call.Fun.(type) {
					case *ast.Ident:
						calleeName = fun.Name
					case *ast.SelectorExpr:
						calleeName = fun.Sel.Name
					}

					if calleeName != "" && !isLowValueCall(calleeName) {
						fileMap.Calls = append(fileMap.Calls, CallRelation{
							From: funcName,
							To:   calleeName,
						})
					}
					return true
				})
			}

		case *ast.GenDecl:
			for _, spec := range d.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				switch typeSpec.Type.(type) {
				case *ast.StructType:
					fileMap.Structs = append(
						fileMap.Structs,
						typeSpec.Name.Name,
					)

				case *ast.InterfaceType:
					fileMap.Interfaces = append(
						fileMap.Interfaces,
						typeSpec.Name.Name,
					)
				}
			}
		}
	}

	return fileMap
}

func parseFile(path string) FileMap {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".go":
		return parseGoFile(path)

	case ".py":
		return parsePythonFile(path)

	case ".js", ".jsx", ".ts", ".tsx":
		return parseJSFile(path)

	case ".java":
		return parseJavaFile(path)

	default:
		return FileMap{
			Path: path,
		}
	}
}

func parsePythonFile(path string) FileMap {
	content, _ := os.ReadFile(path)
	text := string(content)

	fileMap := FileMap{
		Path:     path,
		Language: "python",
	}

	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "import ") || strings.HasPrefix(line, "from ") {
			fileMap.Imports = append(fileMap.Imports, line)
			continue
		}

		// Function detection (including async def)
		if strings.HasPrefix(line, "def ") || strings.HasPrefix(line, "async def ") {
			trimmed := strings.TrimPrefix(line, "async ")
			trimmed = strings.TrimPrefix(trimmed, "def ")
			name := strings.Split(trimmed, "(")[0]
			name = strings.TrimSpace(name)
			if name != "" {
				fileMap.Functions = append(fileMap.Functions, Function{
					Name: name,
				})
			}
			continue
		}

		// Class detection
		if strings.HasPrefix(line, "class ") {
			name := strings.TrimPrefix(line, "class ")
			name = strings.Split(name, "(")[0]
			name = strings.Split(name, ":")[0]
			name = strings.TrimSpace(name)
			if name != "" {
				fileMap.Classes = append(fileMap.Classes, name)
			}
			continue
		}

		// Basic decorator detection (adding to imports or a new field if we had one, but let's just ignore for now or log)
		// User mentioned "decorators awareness"
	}

	return fileMap
}



func parseJSFile(path string) FileMap {
	content, _ := os.ReadFile(path)
	text := string(content)

	fileMap := FileMap{
		Path:     path,
		Language: "javascript",
	}

	lines := strings.Split(text, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "import ") {
			fileMap.Imports = append(fileMap.Imports, line)
			continue
		}

		// function declarations and exports
		if strings.Contains(line, "function ") {
			trimmed := strings.TrimPrefix(line, "export ")
			trimmed = strings.TrimPrefix(trimmed, "default ")
			if strings.HasPrefix(trimmed, "function ") {
				name := strings.Split(strings.TrimPrefix(trimmed, "function "), "(")[0]
				name = strings.TrimSpace(name)
				if name != "" {
					fileMap.Functions = append(fileMap.Functions, Function{
						Name: name,
					})
				}
				continue
			}
		}

		// arrow functions: const name = (...) =>
		if (strings.Contains(line, "const ") || strings.Contains(line, "let ") || strings.Contains(line, "var ")) &&
			strings.Contains(line, "=>") {
			parts := strings.Split(line, "=")
			if len(parts) > 1 {
				decl := strings.Fields(parts[0])
				if len(decl) > 0 {
					name := decl[len(decl)-1]
					fileMap.Functions = append(fileMap.Functions, Function{
						Name: name,
					})
				}
			}
			continue
		}

		// Class detection
		if strings.Contains(line, "class ") {
			trimmed := strings.TrimPrefix(line, "export ")
			trimmed = strings.TrimPrefix(trimmed, "default ")
			if strings.HasPrefix(trimmed, "class ") {
				name := strings.Split(strings.TrimPrefix(trimmed, "class "), " ")[0]
				name = strings.Trim(name, "{")
				name = strings.TrimSpace(name)
				if name != "" {
					fileMap.Classes = append(fileMap.Classes, name)
				}
			}
			continue
		}
	}

	return fileMap
}

func parseJavaFile(path string) FileMap {
	content, _ := os.ReadFile(path)
	text := string(content)

	fileMap := FileMap{
		Path:     path,
		Language: "java",
	}

	lines := strings.Split(text, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// package com.example;
		if strings.HasPrefix(line, "package ") {
			pkg := strings.TrimPrefix(line, "package ")
			pkg = strings.TrimSuffix(pkg, ";")
			fileMap.Package = pkg
		}

		// import java.util.List;
		if strings.HasPrefix(line, "import ") {
			imp := strings.TrimPrefix(line, "import ")
			imp = strings.TrimSuffix(imp, ";")
			fileMap.Imports = append(fileMap.Imports, imp)
		}

		// public class User {
		if strings.Contains(line, " class ") ||
			strings.HasPrefix(line, "class ") {

			parts := strings.Fields(line)

			for idx, part := range parts {
				if part == "class" && idx+1 < len(parts) {
					fileMap.Classes = append(
						fileMap.Classes,
						strings.Trim(parts[idx+1], "{"),
					)
					break
				}
			}
		}

		// public interface UserService {
		if strings.Contains(line, " interface ") ||
			strings.HasPrefix(line, "interface ") {

			parts := strings.Fields(line)

			for idx, part := range parts {
				if part == "interface" && idx+1 < len(parts) {
					fileMap.Interfaces = append(
						fileMap.Interfaces,
						strings.Trim(parts[idx+1], "{"),
					)
					break
				}
			}
		}

		// simple method detection
		// public void login() {
		if strings.Contains(line, "(") &&
			strings.Contains(line, ")") &&
			strings.Contains(line, "{") &&
			!strings.Contains(line, "if") &&
			!strings.Contains(line, "for") &&
			!strings.Contains(line, "while") &&
			!strings.Contains(line, "switch") &&
			!strings.Contains(line, "catch") {

			beforeParen := strings.Split(line, "(")[0]
			parts := strings.Fields(beforeParen)

			if len(parts) > 0 {
				name := parts[len(parts)-1]

				if name != "class" &&
					name != "interface" &&
					name != "new" {

					fileMap.Functions = append(
						fileMap.Functions,
						Function{
							Name: name,
							Line: i + 1,
						},
					)
				}
			}
		}
	}

	return fileMap
}

func isLowValueCall(name string) bool {
	lowValue := map[string]bool{
		"len": true, "append": true, "cap": true, "make": true, "new": true,
		"string": true, "int": true, "int64": true, "float64": true,
		"Println": true, "Printf": true, "Print": true, "Sprintf": true,
		"Error": true, "Errorf": true, "Exit": true, "Fatal": true, "Fatalf": true,
		"Panic": true, "recover": true, "close": true, "delete": true,
		"copy": true, "real": true, "imag": true, "complex": true,
	}
	return lowValue[name]
}

func calculateImportance(f *FileMap, inboundCount int) int {
	if isEntryPoint(f.Path, f) {
		return 100
	}

	score := 10
	switch f.Role {
	case "state-manager", "persistence-layer":
		score = 80
	case "orchestrator", "request-handler":
		score = 60
	case "external-adapter":
		score = 40
	case "config-layer", "ui-layer":
		score = 30
	case "memory-layer":
		score = 70
	}

	// Add dynamic metrics
	score += inboundCount * 5 // Fan-in weight
	score += len(f.ImportantCalls) * 2 // Fan-out weight

	// Boost for key functions
	score += len(f.KeyFunctions) * 2

	// Final caps per tier
	if f.Role == "utility" && score > 30 {
		score = 30
	}
	if score > 95 && !isEntryPoint(f.Path, f) {
		score = 95
	}
	if score < 10 {
		score = 10
	}

	return score
}

func classifyRole(path string, f *FileMap) string {
	if isEntryPoint(path, f) {
		return "entrypoint"
	}

	p := strings.ToLower(path)

	// State and Persistence
	if strings.Contains(p, "session") {
		return "state-manager"
	}
	if strings.Contains(p, "db/") || strings.Contains(p, "database") || strings.Contains(p, "repository") {
		return "persistence-layer"
	}
	if strings.Contains(p, "config") {
		return "config-layer"
	}

	// AI and Memory
	if strings.Contains(p, "compress") || strings.Contains(p, "prompt") || strings.Contains(p, "memory") || strings.Contains(p, "observation") {
		return "memory-layer"
	}
	if strings.Contains(p, "adapter") {
		return "external-adapter"
	}

	// Orchestration and Handling
	if strings.Contains(p, "handler") || strings.Contains(p, "serve") || strings.Contains(p, "mcp") {
		if strings.Contains(p, "handle") {
			return "request-handler"
		}
		return "orchestrator"
	}
	if strings.Contains(p, "worker") || strings.Contains(p, "job") {
		return "worker/background-job"
	}

	// UI and Injection
	if strings.Contains(p, "ui") || strings.Contains(p, "frontend") || strings.Contains(p, "welcome") {
		return "ui-layer"
	}
	if strings.Contains(p, "inject") || strings.Contains(p, "container") {
		return "injector"
	}

	// Logic vs Utility
	if len(f.KeyFunctions) > 2 {
		return "orchestrator"
	}

	return "utility"
}

func isIgnoreZone(path string) bool {
	p := strings.ToLower(path)
	return strings.Contains(p, "_test.go") ||
		strings.Contains(p, "vendor/") ||
		strings.Contains(p, "generated/") ||
		strings.Contains(p, "mock") ||
		strings.Contains(p, "temp") ||
		strings.Contains(p, "debug")
}

func generateExecutionFlows(files []FileMap) []string {
	var flows []string

	// 1. CLI Execution Path
	for _, f := range files {
		if strings.Contains(f.Path, "main.go") && !strings.Contains(f.Path, "test") {
			steps := []string{"CLI Entry"}
			for _, call := range f.ImportantCalls {
				if isHighSignalFunc(call) {
					steps = append(steps, call)
				}
			}
			flows = append(flows, "CLI Pipeline: "+strings.Join(steps, " → "))
		}
	}

	// 2. MCP / Request Path
	for _, f := range files {
		if strings.Contains(f.Path, "mcpServer") && (strings.Contains(f.Path, "mcp") || strings.Contains(f.Path, "handle")) {
			steps := []string{"MCP Request"}
			for _, fn := range f.KeyFunctions {
				if strings.Contains(strings.ToLower(fn), "handle") || strings.Contains(strings.ToLower(fn), "serve") {
					steps = append(steps, fn)
				}
			}
			if len(steps) > 1 {
				flows = append(flows, "MCP Server Flow: "+strings.Join(steps, " → "))
			}
		}
	}

	// 3. Memory / Hook Pipeline
	for _, f := range files {
		if strings.Contains(f.Path, "handleHooks") || strings.Contains(f.Path, "injector") {
			steps := []string{"Session Lifecycle"}
			for _, call := range f.ImportantCalls {
				if strings.Contains(call, "Prompt") || strings.Contains(call, "Memory") || strings.Contains(call, "Session") {
					steps = append(steps, call)
				}
			}
			if len(steps) > 1 {
				flows = append(flows, "Memory Intelligence Flow: "+strings.Join(steps, " → "))
			}
		}
	}

	return flows
}

func isHighSignalFunc(name string) bool {
	if len(name) == 0 || name[0] < 'A' || name[0] > 'Z' {
		return false
	}
	noise := map[string]bool{
		"String": true, "Error": true, "Len": true, "Cap": true,
	}
	return !noise[name]
}

func isEntryPoint(path string, f *FileMap) bool {
	p := strings.ToLower(filepath.Base(path))

	// Go
	if f.Package == "main" || p == "main.go" {
		return true
	}

	// JS / TS
	if strings.HasPrefix(p, "index.") ||
		strings.HasPrefix(p, "server.") ||
		strings.HasPrefix(p, "app.") ||
		strings.HasPrefix(p, "main.") ||
		strings.Contains(p, "mcpserver") ||
		strings.Contains(p, "handlerequest") {
		return true
	}

	// Python
	if p == "main.py" || p == "app.py" || p == "server.py" {
		return true
	}

	// Java
	if p == "main.java" || p == "app.java" {
		return true
	}

	return false
}
