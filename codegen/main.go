package main

import (
	"bytes"
	"embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
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
	FullPkg            string
	CompImport         string
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
	Args      []SelectArg
	EarlyStop bool
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("usage: go generate ecs/codegen package_name component_pkg")
	}

	generatedPackage := os.Args[1]
	componentPkg := os.Args[2]

	genFile, ok := os.LookupEnv("GOFILE")
	if !ok {
		log.Fatal("This program should be run via go generate!")
	}
	genFile, err := filepath.Abs(genFile)
	if err != nil {
		log.Fatal(err)
	}
	systemPkg := filepath.Dir(genFile)

	moduleFileDir := systemPkg
	var modData []byte
	var subpackages []string
	for {
		modData, err = os.ReadFile(filepath.Join(moduleFileDir, "go.mod"))
		if err != nil && os.IsNotExist(err) {
			subpackages = append(subpackages, filepath.Base(moduleFileDir))
			moduleFileDir = filepath.Dir(moduleFileDir)
			continue
		}
		if err != nil {
			log.Fatal(err)
		}
		break
	}
	slices.Reverse(subpackages)
	subpackagePath := strings.Join(subpackages, "/")
	modulePath := ModulePath(modData)

	comps, compMap := findComponents(componentPkg)
	selects := findSelects(systemPkg, compMap)

	context := &Ctx{
		Pkg:                generatedPackage,
		FullPkg:            path.Join(modulePath, subpackagePath, generatedPackage),
		CompImport:         fmt.Sprintf(`import comp "%s"`, path.Join(modulePath, subpackagePath, filepath.Clean(componentPkg))),
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
}

var (
	slashSlash = []byte("//")
	moduleStr  = []byte("module")
)

// ModulePath returns the module path from the gomod file text.
// If it cannot find a module path, it returns an empty string.
// It is tolerant of unrelated problems in the go.mod file.
// From golang.org/x/mod/modfile
func ModulePath(mod []byte) string {
	for len(mod) > 0 {
		line := mod
		mod = nil
		if i := bytes.IndexByte(line, '\n'); i >= 0 {
			line, mod = line[:i], line[i+1:]
		}
		if i := bytes.Index(line, slashSlash); i >= 0 {
			line = line[:i]
		}
		line = bytes.TrimSpace(line)
		if !bytes.HasPrefix(line, moduleStr) {
			continue
		}
		line = line[len(moduleStr):]
		n := len(line)
		line = bytes.TrimSpace(line)
		if len(line) == n || len(line) == 0 {
			continue
		}

		if line[0] == '"' || line[0] == '`' {
			p, err := strconv.Unquote(string(line))
			if err != nil {
				return "" // malformed quoted string or multiline module path
			}
			return p
		}

		return string(line)
	}
	return "" // missing module path
}

type foundSelect struct {
	compNames   []string
	earlyReturn bool
}

func findSelects(path string, compNames map[string]int) []Select {
	fset := token.NewFileSet()
	dir, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	selects := make(map[string]foundSelect)

	// Inject at least one select to avoid unuse import errors in the generated select package
	for firstComponent, _ := range compNames {
		selects[firstComponent] = foundSelect{compNames: []string{firstComponent}, earlyReturn: false}
		break
	}

	for _, pkg := range dir {
		for _, fi := range pkg.Files {
			ast.Inspect(fi, func(n ast.Node) bool {
				funcType, ok := n.(*ast.FuncType)
				if !ok {
					return true
				}

				// See if we have a boolean return value to indicate early stop
				var returns []*ast.Field
				if funcType.Results != nil {
					returns = funcType.Results.List
				}
				if len(returns) > 1 {
					return true
				}

				earlyReturn := false
				for _, ret := range returns {
					ident, ok := ret.Type.(*ast.Ident)
					if !ok {
						return true
					}
					if ident.Name != "bool" {
						return true
					}
					earlyReturn = true
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

				extraKey := ""
				if earlyReturn {
					extraKey = ",early_return"
				}
				selects[strings.Join(comps, ",")+extraKey] = foundSelect{compNames: comps, earlyReturn: earlyReturn}

				return true
			})
		}
	}

	var uniqueSelects []Select
	for _, val := range selects {
		args := make([]SelectArg, 0)
		for _, compName := range val.compNames {
			args = append(args, SelectArg{
				Name:      compName,
				CompIndex: compNames[compName],
			})
		}
		uniqueSelects = append(uniqueSelects, Select{
			Args:      args,
			EarlyStop: val.earlyReturn,
		})
	}
	return uniqueSelects
}

func findComponents(path string) ([]Component, map[string]int) {
	fset := token.NewFileSet()
	dir, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	var components []Component
	for _, pkg := range dir {
		for _, fi := range pkg.Files {
			ast.Inspect(fi, func(n ast.Node) bool {
				typeSpec, ok := n.(*ast.TypeSpec)
				if !ok {
					return true
				}

				components = append(components, Component{typeSpec.Name.Name})
				return true
			})
		}
	}

	compMap := make(map[string]int)
	for i, component := range components {
		compMap[component.Name] = i
	}
	return components, compMap
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
