package commands

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type FileMap struct {
	Path       string   `json:"path"`
	Package    string   `json:"package"`
	Imports    []string `json:"imports"`
	Functions  []string `json:"functions"`
	Structs    []string `json:"structs"`
	Interfaces []string `json:"interfaces"`
	Methods    []Method `json:"methods"`
}

type Method struct {
	Receiver string `json:"receiver"`
	Name     string `json:"name"`
}

type CodebaseMap struct {
	Files []FileMap `json:"files"`
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

			if name == ".git" ||
				name == "vendor" ||
				name == "node_modules" ||
				name == "dist" ||
				name == "build" ||
				name == ".cogito" {
				return filepath.SkipDir
			}

			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		fileMap := parseGoFile(path)
		result.Files = append(result.Files, fileMap)

		return nil
	})

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

	for _, imp := range node.Imports {
		fileMap.Imports = append(
			fileMap.Imports,
			strings.Trim(imp.Path.Value, `"`),
		)
	}

	for _, decl := range node.Decls {
		switch d := decl.(type) {

		case *ast.FuncDecl:
			if d.Recv == nil {
				fileMap.Functions = append(
					fileMap.Functions,
					d.Name.Name,
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
						Name:     d.Name.Name,
					},
				)
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
