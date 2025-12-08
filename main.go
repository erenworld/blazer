package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"log"
	"os"
	"os/exec"
	"slices"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Example of the typed-nil interface problem.
type T struct{ Test string }

func (t *T) Error() string { return t.Test }

func api() *T			   { return nil }

func use() {
	var err error = api()
	if err == nil {
		panic("this is impossible")
	}
}

// Check the AST subtree complexity in a depth-first traversal.
// For each node, it calls the anonymous function.
func checkSubtreeComplexity(node ast.Node) bool {
	isComplex := false

	ast.Inspect(node, func(n ast.Node) bool {
		switch node.(type) {
		case *ast.CallExpr:
			isComplex = true
		case *ast.ReturnStmt:
			isComplex = true
		}
		return true
	})

	return isComplex
}

func main() {
	fmt.Println("Welcome to Blazer linter !")

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	log.Printf("Current Working Directory: %v", cwd)
	cmd := exec.Command("go", "build", "-gcflags", "-S", "./...")
	log.Printf("Compilation for assembly inspection")
	
	go func() { cmd.Start() }()
	stdout, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	assemblyLines := make(map[string][]int)
	scanner := bufio.NewScanner(stdout)

	for scanner.Scan() {
		line := scanner.Text()
		cwdIndex := strings.Index(line, cwd)
		if cwdIndex == -1 {
			continue
		}

		// Extract file path and line number from assembly output.
		colonIndex := strings.Index(line[cwdIndex:], ":")
		if colonIndex == -1 {
			log.Printf("Line has incorrect format")
			continue
		}
 		tabIndex := strings.Index(line[cwdIndex:], "\t")
		if tabIndex == -1 {
			log.Printf("Line has incorrect format")
			continue
		}
		filePath := line[cwdIndex : (cwdIndex + colonIndex)]
		textLine := line[cwdIndex + colonIndex + 1 : (cwdIndex + tabIndex - 1)]
		// log.Printf("FILEPATH=%q  LINETEXT=%q", filePath, textLine)
		lineNumber, err := strconv.Atoi(textLine)
		if err != nil {
			log.Printf("Error: cannot convert line number")
			continue
		}
		assemblyLines[filePath] = append(assemblyLines[filePath], lineNumber)		
	}

	log.Printf("assembly information were indexed, prepare it for queries")

	// line filtering
	for fileName, lineNumbers := range assemblyLines {
		sort.Ints(lineNumbers)
		newSlice := make([]int, 0)
		for i := 0; i < len(lineNumbers); i++ {
			if i == 0 || lineNumbers[i] != lineNumbers[i-1] {
				newSlice = append(newSlice, lineNumbers[i])
			}
		}
		assemblyLines[fileName] = newSlice
	}

	
	log.Printf("assembly information were prepared, ready to process AST")

	// Process AST.
	config := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedFiles | packages.NeedTypes,
		Tests: true,
	}

	project, err := packages.Load(config, "./...")
	if err != nil {
		panic(err)
	}
	for _, pkg := range project {
		for _, file := range pkg.Syntax {
			// Ignore generated files
			filename := pkg.Fset.Position(file.Pos()).Filename
			if strings.Contains(filename, "generated") {
				continue
			}
			var currentFunc string
			ast.Inspect(file, func(node ast.Node) bool {
				if funcDecl, ok := node.(*ast.FuncDecl); ok {
					currentFunc = funcDecl.Name.Name
					return true
				}
				_, ok := node.(*ast.IfStmt)
				if !ok {
					return true
				}
				if !checkSubtreeComplexity(node) {
					return false
				}
				start, end := pkg.Fset.Position(node.Pos()), pkg.Fset.Position(node.End())
				startLine, endLine := start.Line, end.Line

				// Check assembly
				lines, ok := assemblyLines[start.Filename]
				if !ok { 
					return true 
				}
				position, _ := slices.BinarySearch(lines, startLine)
				if position >= len(lines) || lines[position] > endLine {
					log.Printf("The Go Compiler removed this block: func=%s file=%s line=%d\n", currentFunc, start.Filename, start.Line)
				}
				return true
			})
		}
	}
}