package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/annuaicoder/yoft/compiler/codegen"
	"github.com/annuaicoder/yoft/compiler/lexer"
	"github.com/annuaicoder/yoft/compiler/parser"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "version", "--version", "-v":
		fmt.Printf("Yoft %s\n", version)
	case "help", "--help", "-h":
		printUsage()
	case "build":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: yoft build requires a source file")
			fmt.Fprintln(os.Stderr, "Usage: yoft build <file.eng> [-o output]")
			os.Exit(1)
		}
		sourceFile := os.Args[2]
		outputFile := ""
		for i := 3; i < len(os.Args); i++ {
			if os.Args[i] == "-o" && i+1 < len(os.Args) {
				outputFile = os.Args[i+1]
				i++
			}
		}
		if outputFile == "" {
			base := filepath.Base(sourceFile)
			ext := filepath.Ext(base)
			outputFile = strings.TrimSuffix(base, ext)
		}
		err := buildFile(sourceFile, outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "run":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Error: yoft run requires a source file")
			fmt.Fprintln(os.Stderr, "Usage: yoft run <file.eng>")
			os.Exit(1)
		}
		sourceFile := os.Args[2]
		err := runFile(sourceFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		// If arg ends with .eng, treat as "run"
		if strings.HasSuffix(os.Args[1], ".eng") {
			err := runFile(os.Args[1])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
			printUsage()
			os.Exit(1)
		}
	}
}

func printUsage() {
	fmt.Println("Yoft - A compiled programming language")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  yoft build <file.eng> [-o output]   Compile to native binary")
	fmt.Println("  yoft run <file.eng>                  Compile and run immediately")
	fmt.Println("  yoft <file.eng>                      Shorthand for 'yoft run'")
	fmt.Println("  yoft version                         Show version")
	fmt.Println("  yoft help                            Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  yoft build hello.eng -o hello        Compile hello.eng to ./hello")
	fmt.Println("  yoft run hello.eng                   Compile and run hello.eng")
	fmt.Println("  yoft hello.eng                       Same as 'yoft run hello.eng'")
}

func compile(source string) (string, error) {
	lex := lexer.New(source)
	tokens, err := lex.Tokenize()
	if err != nil {
		return "", err
	}

	p := parser.New(tokens)
	program, err := p.Parse()
	if err != nil {
		return "", err
	}

	gen := codegen.New()
	cCode := gen.Generate(program)
	return cCode, nil
}

func buildFile(sourceFile, outputFile string) error {
	source, err := os.ReadFile(sourceFile)
	if err != nil {
		return fmt.Errorf("cannot read '%s': %v", sourceFile, err)
	}

	cCode, err := compile(string(source))
	if err != nil {
		return err
	}

	// Write C file to temp
	tmpC := outputFile + ".c"
	err = os.WriteFile(tmpC, []byte(cCode), 0644)
	if err != nil {
		return fmt.Errorf("cannot write temp C file: %v", err)
	}
	defer os.Remove(tmpC)

	// Compile with cc (system C compiler)
	cmd := exec.Command("cc", "-O2", "-o", outputFile, tmpC, "-lm")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("C compilation failed: %v", err)
	}

	fmt.Printf("Compiled: %s -> %s\n", sourceFile, outputFile)
	return nil
}

func runFile(sourceFile string) error {
	// Build to temp
	tmpDir, err := os.MkdirTemp("", "yoft-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	tmpBin := filepath.Join(tmpDir, "yoft_out")
	err = buildFile(sourceFile, tmpBin)
	if err != nil {
		return err
	}

	// Execute
	cmd := exec.Command(tmpBin)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
