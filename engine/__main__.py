#!/usr/bin/env python3
"""
Command-line entry point for the Engine Language interpreter.
"""

import sys
from .interpreter import run_file


def main():
    """Main entry point for the engine command."""
    if len(sys.argv) != 2:
        print("Usage: engine <filename.eng>")
        print("")
        print("Run Engine language files (.eng) through the interpreter.")
        print("")
        print("Example:")
        print("  engine my_program.eng")
        sys.exit(1)
    
    filename = sys.argv[1]
    if not filename.endswith('.eng'):
        print(f"Warning: '{filename}' does not have a .eng extension.")
    
    run_file(filename)


if __name__ == "__main__":
    main()
