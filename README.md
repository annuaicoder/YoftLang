# Engine-Lang

# Built by @annuaicoder

# Link to Youtube Cha

**Engine** is a minimal, readable interpreted programming language written in Python. It is actively maintained and under rapid development — expect bugs and breaking changes along the way.

---

## Install

Install Engine Lang directly from GitHub:

```bash
pip install git+https://github.com/annuaicoder/EngineLang.git

## Quick Start

```bash
# Run a script
python -m engine test.eng

# Or if installed via pip:
engine test.eng

# Start interactive REPL
python -m engine
```

### Install (editable / development)

```bash
pip install -e .
```

---

## Language Overview

### Variables

```
var name = "Engine"
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
items.sort()
```

List methods: `push`, `pop`, `join`, `contains`, `reverse`, `sort`, `slice`, `length`.

### Strings

```
var s = "hello"
show s.upper()
show s.contains("ell")
show s.split("l")
show s.replace("l", "r")
```

String methods: `upper`, `lower`, `strip`, `split`, `replace`, `contains`, `starts_with`, `ends_with`, `find`, `length`.

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
| `min(...)` | Minimum value |
| `max(...)` | Maximum value |
| `round(x)` | Round a number |

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

See the `examples/` folder:

- `hello.eng` — Hello World
- `variables.eng` — Variables and expressions
- `control_flow.eng` — If/else, while, for, booleans
- `functions.eng` — Functions, recursion, closures
- `lists_strings.eng` — Lists and string operations
- `fizzbuzz.eng` — Classic FizzBuzz

Run any example:

```bash
python -m engine examples/fizzbuzz.eng
```

---

## Project Structure

```
Engine-Lang/
  engine/
    __init__.py        # Package exports
    __main__.py        # CLI entry point & REPL
    interpreter.py     # Lexer, parser, AST, interpreter
  examples/            # Example .eng scripts
  quickstart/          # Beginner docs
  test.eng             # Quick test script
  setup.py             # Package setup
  pyproject.toml       # Build config
```

---

Documentation: https://github.com/annuaicoder/Engine-Lang

Pull requests are welcome and encouraged!

