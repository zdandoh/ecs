package main

import (
	"bytes"
	"embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed ecs
var fs embed.FS

var templFuncs = map[string]interface{}{
	"lshift": func(x int) int {
		return 1 << x
	},
	"compmapindex": func(index int) int {
		return index / 64
	},
	"compsubindex": func(index int) int {
		return 1 << (index % 64)
	},
}

type Ctx struct {
	Pkg                string
	Comps              []Component
	CompCount          int
	CompContainerCount int
	Selects            []Select
	SelectCount        int
}

type Component struct {
	Name string
}

type SelectArg struct {
	Name      string
	CompIndex int
}

type Select struct {
	Args []SelectArg
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("usage: go generate ecs/codegen package_name component_defs")
	}

	generatedPackage := os.Args[1]
	componentFile := os.Args[2]

	genFile, ok := os.LookupEnv("GOFILE")
	if !ok {
		log.Fatal("This program should be run via go generate!")
	}
	genFile, err := filepath.Abs(genFile)
	if err != nil {
		log.Fatal(err)
	}
	systemPkg := filepath.Dir(genFile)

	comps, compMap, compFile := findComponents(componentFile, generatedPackage)
	selects := findSelects(systemPkg, compMap)

	context := &Ctx{
		Pkg:                generatedPackage,
		Comps:              comps,
		CompCount:          len(comps),
		CompContainerCount: int(math.Ceil(float64(len(comps)) / 64)),
		Selects:            selects,
		SelectCount:        len(selects),
	}

	err = setupPackage(context)
	if err != nil {
		log.Fatal(fmt.Errorf("error setting up package: %w", err))
	}

	err = ioutil.WriteFile(filepath.Join(generatedPackage, "gen_components.go"), []byte(compFile), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func updatePackage(generatedPackage string, fi *ast.File, fset *token.FileSet) string {
	fi.Name.Name = generatedPackage
	buf := bytes.NewBuffer(nil)
	printer.Fprint(buf, fset, fi)
	return buf.String()
}

func findSelects(path string, compNames map[string]int) []Select {
	fset := token.NewFileSet()
	dir, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	selects := make(map[string][]string)
	for _, pkg := range dir {
		for _, fi := range pkg.Files {
			ast.Inspect(fi, func(n ast.Node) bool {
				funcType, ok := n.(*ast.FuncType)
				if !ok {
					return true
				}

				params := funcType.Params.List
				if len(params) < 2 {
					return true
				}
				if entityParam, ok := params[0].Type.(*ast.SelectorExpr); ok {
					if entityParam.Sel.Name != "Entity" {
						return true
					}
				} else {
					return true
				}

				var comps []string
				for _, param := range funcType.Params.List[1:] {
					switch paramT := param.Type.(type) {
					case *ast.Ident:
						return true
					case *ast.SelectorExpr:
						return true
					case *ast.StarExpr:
						switch startT := paramT.X.(type) {
						case *ast.Ident:
							return true
						case *ast.SelectorExpr:
							_, ok := compNames[startT.Sel.Name]
							if !ok {
								return true
							}
							comps = append(comps, startT.Sel.Name)
						}
					}
				}
				selects[strings.Join(comps, ",")] = comps

				return true
			})
		}
	}

	var uniqueSelects []Select
	for _, val := range selects {
		args := make([]SelectArg, 0)
		for _, compName := range val {
			args = append(args, SelectArg{
				Name:      compName,
				CompIndex: compNames[compName],
			})
		}
		uniqueSelects = append(uniqueSelects, Select{
			Args: args,
		})
	}
	return uniqueSelects
}

func findComponents(path string, newPkg string) ([]Component, map[string]int, string) {
	fset := token.NewFileSet()
	fi, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	var components []Component
	ast.Inspect(fi, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		components = append(components, Component{typeSpec.Name.Name})
		return true
	})

	compMap := make(map[string]int)
	for i, component := range components {
		compMap[component.Name] = i
	}
	return components, compMap, updatePackage(newPkg, fi, fset)
}

func recursiveCopy(fs embed.FS, dir string, packageName string, context *Ctx) error {
	err := os.Mkdir(packageName, 0744)
	if err != nil {
		return err
	}

	files, err := fs.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("readdir err: %w", err)
	}

	for _, fi := range files {
		if fi.IsDir() {
			err := recursiveCopy(fs, dir+"/"+fi.Name(), filepath.Join(packageName, fi.Name()), context)
			if err != nil {
				return err
			}
			continue
		}
		sourceFi, err := fs.Open(dir + "/" + fi.Name())
		if err != nil {
			return err
		}
		defer sourceFi.Close()
		newFi, err := os.Create(filepath.Join(packageName, strings.Replace(fi.Name(), ".go.templ", ".go", -1)))
		if err != nil {
			return err
		}
		defer newFi.Close()

		sourceBytes, err := ioutil.ReadAll(sourceFi)
		if err != nil {
			return err
		}

		if strings.HasSuffix(fi.Name(), ".go.templ") {
			t, err := template.New(fi.Name()).Funcs(templFuncs).Parse(string(sourceBytes))
			if err != nil {
				return err
			}
			err = t.Execute(newFi, context)
			if err != nil {
				return err
			}
		} else {
			_, _ = newFi.Write(sourceBytes)
		}
	}

	return nil
}

func setupPackage(context *Ctx) error {
	_ = os.RemoveAll(context.Pkg)

	err := recursiveCopy(fs, "ecs", context.Pkg, context)
	if err != nil {
		return err
	}

	return nil
}
