package main

import (
	"bytes"
	"embed"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

//go:embed ecs
var fs embed.FS

type Template struct {
	Header   string
	Repeated string
	Footer   string
}

var templates = []Template{
	{
		Header: `const MaxEntities = %MAX_ENTITIES%
var ranges [%COMPONENT_COUNT%]componentRange`,
	},
	{ // Define component storage
		Repeated: `
var store%s []%s`,
	},
	{ // Allocate component storage
		Header: `
func init() {
`,
		Repeated: `   store%s = make([]%s, MaxEntities)`,
		Footer: `
}`,
	},
	{ // Reset component ranges
		Header: `
func init() {
    for i := 0; i < %COMPONENT_COUNT%; i++ {
		ranges[i].Reset()
	}
}
`,
	},
	{ // Clear global component state
		Header:   "\nfunc clearComponents() {\n",
		Repeated: "    store%s = make([]%s, MaxEntities)",
		Footer:   "}\n",
	},
	{ // Component ID
		Repeated: `
const Component%s ComponentID = %COMPONENT_POW%`,
	},
	{ // Add component
		Repeated: `
func (e *Entity) Add%s(c %s) {
	e.components |= Component%s
	ranges[%COMPONENT_INDEX%].Add(e.id)
	store%s[e.id] = c
}
`,
	},
	{ // Remove component
		Repeated: `
func (e *Entity) Remove%s() {
	e.components &= ^Component%s
}`,
	},
	{ // Get component
		Repeated: `
func (e *Entity) %s() *%s {
	return &store%s[e.id]
}`,
	},
}

type Component struct {
	Name string
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("usage: go generate ecs/codegen package_name max_entites")
	}

	generatedPackage := os.Args[1]

	maxEntities, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal("couldn't parse max entity count: %v", err)
	}

	inputFi := os.Getenv("GOFILE")

	fset := token.NewFileSet()
	fi, err := parser.ParseFile(fset, inputFi, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	err = setupPackage(generatedPackage)
	if err != nil {
		log.Fatal(fmt.Errorf("error setting up package: %w", err))
	}

	components := findComponents(fi)
	newSrc := updatePackage(generatedPackage, fi, fset)
	outFi := genCode(components, maxEntities, newSrc)

	err = ioutil.WriteFile(filepath.Join(generatedPackage, "gen_"+inputFi), []byte(outFi), 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func formatTemplate(template string, maxEntities int, curr Component, totalComps int, index int) string {
	t1 := strings.ReplaceAll(template, "%s", curr.Name)
	t2 := strings.ReplaceAll(t1, "%COMPONENT_INDEX%", strconv.Itoa(index))
	t3 := strings.ReplaceAll(t2, "%COMPONENT_COUNT%", strconv.Itoa(totalComps))
	t4 := strings.ReplaceAll(t3, "%MAX_ENTITIES%", strconv.Itoa(maxEntities))
	t5 := strings.ReplaceAll(t4, "%COMPONENT_POW%", strconv.Itoa(1<<index))
	return t5
}

func genCode(components []Component, maxEntities int, sourceFile string) string {
	outputFi := strings.Builder{}

	outputFi.WriteString(sourceFile + "\n")

	for _, temp := range templates {
		outputFi.WriteString(formatTemplate(temp.Header, maxEntities, Component{}, len(components), 0))
		for i, comp := range components {
			outputFi.WriteString(formatTemplate(temp.Repeated, maxEntities, comp, len(components), i) + "\n")
		}
		outputFi.WriteString(formatTemplate(temp.Footer, maxEntities, Component{}, len(components), 0))
	}

	outStr := outputFi.String()
	fmt.Println(outStr)
	return outStr
}

func updatePackage(generatedPackage string, fi *ast.File, fset *token.FileSet) string {
	fi.Name.Name = generatedPackage
	buf := bytes.NewBuffer(nil)
	printer.Fprint(buf, fset, fi)
	return buf.String()
}

func findComponents(fi *ast.File) []Component {
	var components []Component

	ast.Inspect(fi, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		components = append(components, Component{typeSpec.Name.Name})
		return true
	})

	return components
}

func setupPackage(generatedPackage string) error {
	_ = os.RemoveAll(generatedPackage)

	files, err := fs.ReadDir("ecs")
	if err != nil {
		return err
	}

	err = os.Mkdir(generatedPackage, os.ModeDir)
	if err != nil {
		return err
	}

	for _, fi := range files {
		sourceFi, err := fs.Open("ecs/" + fi.Name())
		if err != nil {
			return err
		}
		defer sourceFi.Close()
		newFi, err := os.Create(filepath.Join(generatedPackage, fi.Name()))
		if err != nil {
			return err
		}
		defer newFi.Close()

		_, err = io.Copy(newFi, sourceFi)
		if err != nil {
			return err
		}
	}

	return nil
}
