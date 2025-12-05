package main

import (
	"bufio"
	"fmt"
	"go/ast"
	"os/exec"
	"os"
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
		// A function call almost always has some effect.
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
}