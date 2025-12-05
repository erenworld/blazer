## ⚡️ Blazer
Go linter that reports code removed by the Go compiler’s optimizations.

Blazer analyzes the compiler’s assembly output to identify source lines that are eliminated during compilation. Removed lines often indicate constant conditions, unreachable code, or incorrect handling of typed nil values in interfaces.

## Why
Go may remove entire branches when it can statically determine that a condition is always true or always false. This frequently happens with typed nil values returned through interfaces, leading to subtle logical issues. Blazer highlights these cases by detecting missing source lines in the final assembly.

A source line disappears from assembly when the Go optimizer proves that the code:
- does nothing
- is unreachable
- is a dead conditional branch
- is always replaced by a constant
- or is simplified away entirely

## How Blazer Works
1. Compile the file with `-gcflags=-S` to obtain the assembly output.
2. Extract the line numbers that appear in the compiler output.
3. Compare them with the original source file.
4. Any source line that does not appear in the assembly has been removed by optimization.

## Typed Nil and Interfaces
A Go interface holds two fields:
- a concrete type
- a value

An interface is only equal to nil when both fields are nil. Returning a typed nil (for example, a nil pointer to a concrete type) produces an interface value with:
type = non-nil
value = nil

This interface is not equal to nil, even though the underlying pointer is nil.

## Example of the Issue

```go
package main

import "net/url"

func makeNilPtr() *url.Error { return nil }
func asInterface() error      { return makeNilPtr() }

func main() {
    if err := asInterface(); err == nil {
        panic("never called")
    }
}
```

The function asInterface returns a non-nil interface. The compiler detects that the condition err == nil is always false and removes the entire branch. The corresponding line will not appear in the assembly output.

## Typical Assembly Check

`go test -c -gcflags=-S main.go 2>&1 | grep -Eo "main.go:[0-9]+" | uniq`

If a line is missing (for example the line containing the if condition), this indicates that the compiler optimized it away.

## Why It Matters
Code removal indicates that the compiler has proven a branch to be irrelevant or unreachable. In cases involving typed nil values, this often reveals a logical error:

- err != nil is true when the developer expects false
- err == nil is false when the developer expects true

Blazer detects these situations by identifying which lines were eliminated during optimization.
