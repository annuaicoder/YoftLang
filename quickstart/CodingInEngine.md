# Coding in Engine

## Hello World

Use `show` to print to the console:

```
show "Hello World"
```

## Variables

Declare with `var`, reassign with `=`:

```
var x = 5
var name = "Engine"
show x
show name

x = 10
show x
```

## Expressions

Arithmetic: `+`, `-`, `*`, `/`, `%`

```
show 2 + 3 * 4
show (2 + 3) * 4
```

Comparisons: `==`, `!=`, `<`, `>`, `<=`, `>=`

Logic: `and`, `or`, `not`

## If / Else

```
var age = 18

if age >= 18 {
    show "Adult"
} else {
    show "Minor"
}
```

## Loops

```
# While loop
var i = 0
while i < 3 {
    show i
    i = i + 1
}

# For loop
for n in range(5) {
    show n
}
```

## Functions

```
func greet(name) {
    show "Hello, " + name + "!"
}

greet("Alice")

func add(a, b) {
    return a + b
}

show add(3, 7)
```

## Lists

```
var colors = ["red", "green", "blue"]
show colors[0]
colors.push("yellow")
show len(colors)

for c in colors {
    show c
}
```

## String Methods

```
var s = "hello"
show s.upper()
show s.contains("ell")
show s.replace("l", "r")
```

## Comments

```
# This is a comment
var x = 10  # inline comment
```

## Running Scripts

Save your code in a `.eng` file and run:

```bash
python -m engine my_script.eng
```

Or start the interactive REPL:

```bash
python -m engine
```
