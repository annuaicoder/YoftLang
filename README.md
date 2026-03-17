<<<<<<< HEAD
# Engine-Lang - Minimal - Fast and Extremely Easy

# Built by @annuaicoder

## Link to Youtube Channel / Handle: @codewithanas007
=======
# YoftLang
>>>>>>> a3020a6 (Full Rebrand Change To Yoft, Make Language Compiled, Rewrite Compiler In Go, Full Redesign)

**Yoft** is a fast, compiled programming language that produces native binaries. It features clean syntax, zero runtime dependencies, and compiles through C for maximum performance.

> Formerly known as **Engine** — rewritten from the ground up as a compiled language.

---

## Install

### From Source (requires Go 1.21+ and a C compiler)

```bash
git clone https://github.com/annuaicoder/Engine-Lang.git
cd Engine-Lang
go build -o yoft .
```

Move to your PATH:

```bash
sudo mv yoft /usr/local/bin/
```

---

## Quick Start

```bash
# Compile and run immediately
yoft run hello.eng

<<<<<<< HEAD
# Or
engine test.eng
=======
# Compile to native binary
yoft build hello.eng -o hello
>>>>>>> a3020a6 (Full Rebrand Change To Yoft, Make Language Compiled, Rewrite Compiler In Go, Full Redesign)

# Run the compiled binary directly
./hello

# Shorthand: just pass the file
yoft hello.eng
```

---

## How It Works

Yoft is a **true compiled language**:

1. Your `.eng` source is parsed by the Yoft compiler (written in Go)
2. The compiler generates optimized C code
3. The system C compiler (`cc`) produces a native binary
4. The result is a standalone executable — no runtime, no VM, no dependencies

---

## Language Overview

### Variables

```
var name = "Yoft"
var x = 10
var pi = 3.14
var active = true
```

Reassign with just `=`:

```
x = 42
```

### Output

```
show "Hello World"
show x + y
```

### Arithmetic & Comparisons

`+`, `-`, `*`, `/`, `%` for math. `/` does integer division when both operands are ints.

`==`, `!=`, `<`, `>`, `<=`, `>=` for comparisons.

`and`, `or`, `not` for logic.

### If / Else

```
if score >= 90 {
    show "A"
} else if score >= 80 {
    show "B"
} else {
    show "C"
}
```

### While Loop

```
var i = 0
while i < 5 {
    show i
    i = i + 1
}
```

### For Loop

```
for i in range(10) {
    show i
}

for item in my_list {
    show item
}
```

### Functions

```
func greet(name) {
    show "Hello, " + name + "!"
}

func add(a, b) {
    return a + b
}

greet("Alice")
show add(3, 7)
```

Functions support local scope, closures, and recursion.

### Lists

```
var items = [1, 2, 3]
show items[0]
items.push(4)
show len(items)
```

List methods: `push`, `pop`, `join`, `contains`, `reverse`, `length`.

### Strings

```
var s = "hello"
show s.upper()
show s.contains("ell")
```

String methods: `upper`, `lower`, `contains`, `length`.

### Built-in Functions

| Function | Description |
|----------|-------------|
| `show(x)` | Print a value |
| `len(x)` | Length of string or list |
| `type(x)` | Type name as string |
| `int(x)` | Convert to integer |
| `float(x)` | Convert to float |
| `str(x)` | Convert to string |
| `input(prompt)` | Read user input |
| `range(n)` | List from 0 to n-1 |
| `push(list, item)` | Append to list |
| `pop(list)` | Remove last item |
| `abs(x)` | Absolute value |
| `min(a, b)` | Minimum of two values |
| `max(a, b)` | Maximum of two values |
| `round(x)` | Round a number |
| `rand(a, b)` | Random integer in [a, b] |

### Imports

```
import "math_utils.eng"
```

### Comments

```
# This is a comment
var x = 10  # inline comment
```

---

## Examples

See the `examples/` folder and `demo.eng`:

```bash
yoft run demo.eng
yoft run examples/fizzbuzz.eng
```

---

## Project Structure

```
YoftLang/
  main.go                  # CLI entry point
  compiler/
    lexer/lexer.go         # Tokenizer
    parser/parser.go       # Recursive descent parser
    ast/ast.go             # AST node definitions
    codegen/codegen.go     # C code generator
  examples/                # Example .eng scripts
  docs/                    # Website and documentation
  demo.eng                 # Demo script
  go.mod                   # Go module
```

---

## CLI Reference

```
yoft build <file.eng> [-o output]   Compile to native binary
yoft run <file.eng>                  Compile and run immediately
yoft <file.eng>                      Shorthand for 'yoft run'
yoft version                         Show version
yoft help                            Show help
```

---

Website & Docs: https://annuaicoder.github.io/Engine-Lang/

Pull requests are welcome and encouraged!

