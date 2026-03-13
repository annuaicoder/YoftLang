#!/usr/bin/env python3
"""
Command-line entry point for the Engine Language interpreter.

Usage:
    engine <filename.eng>   — run a script
    engine                  — start interactive REPL
"""

import sys
from .interpreter import (
    run_file, tokenize, Parser, Interpreter, Environment, EngineError
)


def repl():
    """Start an interactive Engine REPL."""
    print("Engine Language v0.1.0  —  type 'exit' to quit")
    env = Environment()
    interp = Interpreter(env=env)

    while True:
        try:
            line = input(">> ")
        except (EOFError, KeyboardInterrupt):
            print("\nBye!")
            break

        if line.strip() in ("exit", "quit"):
            break
        if not line.strip():
            continue

        # Allow multi-line input when braces are unbalanced
        while line.count("{") > line.count("}"):
            try:
                line += "\n" + input(".. ")
            except (EOFError, KeyboardInterrupt):
                print("\nBye!")
                return

        try:
            tokens = tokenize(line)
            parser = Parser(tokens)
            tree = parser.parse()
            interp.visit(tree, env)
        except EngineError as e:
            print(f"Engine Error: {e}")
        except SystemExit:
            pass


def main():
    """Main entry point for the engine command."""
    if len(sys.argv) < 2:
        repl()
        return

    if sys.argv[1] in ("-h", "--help"):
        print("Engine Language v0.1.0")
        print("")
        print("Usage:")
        print("  engine <filename.eng>   Run an Engine script")
        print("  engine                  Start interactive REPL")
        print("")
        print("Examples:")
        print("  engine my_program.eng")
        print("  python -m engine test.eng")
        return

    filename = sys.argv[1]
    if not filename.endswith('.eng'):
        print(f"Warning: '{filename}' does not have a .eng extension.")

    run_file(filename)


if __name__ == "__main__":
    main()
