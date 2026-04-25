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

// Schema v2.0 - Graph-based Intelligence
type Node struct {
	ID           string   `json:"id"`
	Importance   int      `json:"importance"`
	Tags         []string `json:"tags,omitempty"`
	Summary      string   `json:"summary,omitempty"`
	KeyFunctions []string `json:"key_functions,omitempty"`
}

type Edge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Weight int    `json:"weight,omitempty"`
}

type CodebaseMap struct {
	Version     string            `json:"version"`
	Entrypoints []string          `json:"entrypoints,omitempty"`
	StateMap    map[string]string `json:"state_map,omitempty"`
	Nodes       []Node            `json:"nodes,omitempty"`
	Edges       []Edge            `json:"edges,omitempty"`
	Flows       []string          `json:"flows,omitempty"`
}

func BuildMap() {
	root := "."
	_ = LoadCache()
	newCache := Cache{Files: make(map[string]string)}
	var allFiles []FileMap

	fmt.Println("Building codebase intelligence map (v2.0)...")

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return nil
		}
		if info.IsDir() {
			if ShouldSkipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if !isSupported(ext) {
			return nil
		}

		hash := GetFileHash(path)
		newCache.Files[path] = hash

		// Incremental caching logic would reside here
		fileMap := parseFile(path)
		allFiles = append(allFiles, fileMap)
		return nil
	})
	SaveCache(newCache)

	// Pass 1: Symbol indexing
	funcToPath := make(map[string]string)
	for _, f := range allFiles {
		for _, fn := range f.Functions {
			funcToPath[fn.Name] = f.Path
		}
	}

	// Pass 2: Graph construction & Weighting
	edgeMap := make(map[string]int)
	for _, f := range allFiles {
		for _, call := range f.Calls {
			targetPath, ok := funcToPath[call.To]
			if ok && targetPath != f.Path {
				edgeMap[f.Path+"|"+targetPath]++
			}
		}
	}

	// Pass 3: Iterative Importance Propagation
	importance := make(map[string]int)
	for _, f := range allFiles {
		importance[f.Path] = calculateBaseImportance(&f)
	}
	for i := 0; i < 2; i++ {
		nextImportance := make(map[string]int)
		for p, s := range importance {
			nextImportance[p] = s
		}
		for edge, weight := range edgeMap {
			parts := strings.Split(edge, "|")
			from, to := parts[0], parts[1]
			nextImportance[to] += (importance[from] * weight) / 10
		}
		importance = nextImportance
	}

	// Pass 4: Final Assembly
	var nodes []Node
	stateMap := make(map[string]string)
	entrypoints := []string{}

	for _, f := range allFiles {
		score := importance[f.Path]
		if isEntryPoint(f.Path, &f) {
			score = 100
			entrypoints = append(entrypoints, f.Path)
		}
		if score > 100 {
			score = 100
		}

		tags := classifyTags(f.Path, &f)
		node := Node{
			ID:           f.Path,
			Importance:   score,
			Tags:         tags,
			Summary:      f.Summary,
			KeyFunctions: extractKeyFunctions(&f),
		}
		nodes = append(nodes, node)

		// Map state owners
		for _, t := range tags {
			if t == "state-owner" || t == "persistence-layer" {
				key := "data"
				if strings.Contains(f.Path, "session") {
					key = "sessions"
				} else if strings.Contains(f.Path, "memory") || strings.Contains(f.Path, "observation") {
					key = "memory"
				}
				stateMap[key] = f.Path
			}
		}
	}

	var finalEdges []Edge
	for edge, weight := range edgeMap {
		parts := strings.Split(edge, "|")
		finalEdges = append(finalEdges, Edge{From: parts[0], To: parts[1], Weight: weight})
	}

	result := CodebaseMap{
		Version:     "2.0",
		Entrypoints: entrypoints,
		StateMap:    stateMap,
		Nodes:       nodes,
		Edges:       finalEdges,
		Flows:       generateExecutionFlowsFromNodes(nodes, finalEdges),
	}

	sort.Slice(result.Nodes, func(i, j int) bool {
		return result.Nodes[i].Importance > result.Nodes[j].Importance
	})

	os.MkdirAll(".cogito", os.ModePerm)
	file, _ := os.Create(".cogito/map.json")
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	enc.Encode(result)
	fmt.Println("Map successfully created at .cogito/map.json")
}

// Helpers

func isSupported(ext string) bool {
	exts := map[string]bool{".go": true, ".py": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true, ".java": true}
	return exts[ext]
}

func calculateBaseImportance(f *FileMap) int {
	score := 10 + len(f.Functions) + len(f.Structs)*2
	if score > 80 {
		score = 80
	}
	return score
}

func classifyTags(path string, f *FileMap) []string {
	var tags []string
	if isEntryPoint(path, f) {
		tags = append(tags, "entrypoint")
	}
	p := strings.ToLower(path)
	if strings.Contains(p, "db/") || strings.Contains(p, "database") {
		tags = append(tags, "persistence-layer")
	}
	if strings.Contains(p, "session") || strings.Contains(p, "config") {
		tags = append(tags, "state-owner")
	}
	if strings.Contains(p, "adapter") || strings.Contains(p, "api") {
		tags = append(tags, "boundary", "adapter")
	}
	if strings.Contains(p, "handler") || strings.Contains(p, "serve") {
		tags = append(tags, "orchestrator")
	}
	if len(tags) == 0 {
		tags = append(tags, "utility")
	}
	return tags
}

func extractKeyFunctions(f *FileMap) []string {
	var keys []string
	for _, fn := range f.Functions {
		if isHighSignalFunc(fn.Name) {
			keys = append(keys, fn.Name)
		}
	}
	return keys
}

func generateExecutionFlowsFromNodes(nodes []Node, edges []Edge) []string {
	return []string{"Graph-based flow construction enabled (v2.0)"}
}

// Parsers

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
		return FileMap{Path: path}
	}
}

func parseGoFile(path string) FileMap {
	fset := token.NewFileSet()
	fmt.Println("Parsing file:", path)
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return FileMap{Path: path}
	}
	fileMap := FileMap{Path: path, Package: node.Name.Name}
	if node.Doc != nil {
		fileMap.Summary = strings.TrimSpace(node.Doc.Text())
	}
	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			funcName := d.Name.Name
			if d.Recv == nil {
				fileMap.Functions = append(fileMap.Functions, Function{Name: funcName, Line: fset.Position(d.Pos()).Line})
			}
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
						fileMap.Calls = append(fileMap.Calls, CallRelation{From: funcName, To: calleeName})
					}
					return true
				})
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if _, ok := typeSpec.Type.(*ast.StructType); ok {
						fileMap.Structs = append(fileMap.Structs, typeSpec.Name.Name)
					}
				}
			}
		}
	}
	return fileMap
}

func parsePythonFile(path string) FileMap {
	fmt.Println("Parsing file:", path)
	content, _ := os.ReadFile(path)
	text := string(content)
	fileMap := FileMap{Path: path, Language: "python"}
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "def ") || strings.HasPrefix(line, "async def ") {
			trimmed := strings.TrimPrefix(line, "async ")
			trimmed = strings.TrimPrefix(trimmed, "def ")
			name := strings.Split(trimmed, "(")[0]
			if name = strings.TrimSpace(name); name != "" {
				fileMap.Functions = append(fileMap.Functions, Function{Name: name})
			}
		}
		if strings.HasPrefix(line, "class ") {
			name := strings.TrimPrefix(line, "class ")
			name = strings.Split(name, "(")[0]
			name = strings.Split(name, ":")[0]
			if name = strings.TrimSpace(name); name != "" {
				fileMap.Classes = append(fileMap.Classes, name)
			}
		}
	}
	return fileMap
}

func parseJSFile(path string) FileMap {
	fmt.Println("Parsing file:", path)
	content, _ := os.ReadFile(path)
	text := string(content)
	fileMap := FileMap{Path: path, Language: "javascript"}
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "function ") {
			trimmed := strings.TrimPrefix(line, "export ")
			trimmed = strings.TrimPrefix(trimmed, "default ")
			if strings.HasPrefix(trimmed, "function ") {
				name := strings.Split(strings.TrimPrefix(trimmed, "function "), "(")[0]
				if name = strings.TrimSpace(name); name != "" {
					fileMap.Functions = append(fileMap.Functions, Function{Name: name})
				}
			}
		}
		if strings.Contains(line, "class ") {
			trimmed := strings.TrimPrefix(line, "export ")
			trimmed = strings.TrimPrefix(trimmed, "default ")
			if strings.HasPrefix(trimmed, "class ") {
				name := strings.Split(strings.TrimPrefix(trimmed, "class "), " ")[0]
				if name = strings.TrimSpace(strings.Trim(name, "{")); name != "" {
					fileMap.Classes = append(fileMap.Classes, name)
				}
			}
		}
	}
	return fileMap
}

func parseJavaFile(path string) FileMap {
	fmt.Println("Parsing file:", path)
	content, _ := os.ReadFile(path)
	text := string(content)
	fileMap := FileMap{Path: path, Language: "java"}
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "package ") {
			fileMap.Package = strings.TrimSuffix(strings.TrimPrefix(line, "package "), ";")
		}
		if (strings.Contains(line, " class ") || strings.HasPrefix(line, "class ")) && strings.Contains(line, "{") {
			parts := strings.Fields(line)
			for i, p := range parts {
				if p == "class" && i+1 < len(parts) {
					fileMap.Classes = append(fileMap.Classes, strings.Trim(parts[i+1], "{"))
				}
			}
		}
	}
	return fileMap
}

func isEntryPoint(path string, f *FileMap) bool {
	p := strings.ToLower(filepath.Base(path))
	return f.Package == "main" || p == "main.go" || strings.HasPrefix(p, "index.") || strings.HasPrefix(p, "server.") || strings.HasPrefix(p, "app.") || strings.Contains(p, "mcpserver") || strings.Contains(p, "handlerequest")
}

func isHighSignalFunc(name string) bool {
	if len(name) == 0 || name[0] < 'A' || name[0] > 'Z' { return false }
	noise := map[string]bool{"String": true, "Error": true, "Len": true}
	return !noise[name]
}

func isLowValueCall(name string) bool {
	low := map[string]bool{"len": true, "append": true, "make": true, "new": true, "Println": true, "Printf": true}
	return low[name]
}

type FileMap struct {
	Path      string
	Summary   string
	Package   string
	Functions []Function
	Structs   []string
	Classes   []string
	Calls     []CallRelation
	Language  string
}

type Function struct {
	Name string
	Line int
}

type CallRelation struct {
	From string
	To   string
}
