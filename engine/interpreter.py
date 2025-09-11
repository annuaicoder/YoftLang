import re
import sys
import os

# --- CUSTOM ERRORS ---
class EngineError(Exception):
    """Base class for all Engine language errors."""
    pass

class EngineSyntaxError(EngineError):
    def __init__(self, message, line_num=None, line=None):
        self.message = message
        self.line_num = line_num
        self.line = line
        super().__init__(self.__str__())

    def __str__(self):
        if self.line_num is not None and self.line is not None:
            return (f"ENGINE-INTERPRETER: ERROR\n"
                    f"[SyntaxError] Line {self.line_num}: {self.message}\n"
                    f"    {self.line}")
        return f"ENGINE-INTERPRETER: ERROR [SyntaxError] {self.message}"

class EngineNameError(EngineError):
    def __init__(self, name, line_num=None, line=None):
        self.name = name
        self.line_num = line_num
        self.line = line
        super().__init__(self.__str__())

    def __str__(self):
        if self.line_num is not None and self.line is not None:
            return (f"ENGINE-INTERPRETER: ERROR\n"
                    f"[NameError] Line {self.line_num}: Variable '{self.name}' not defined\n"
                    f"    {self.line}")
        return f"ENGINE-INTERPRETER: ERROR [NameError] Variable '{self.name}' not defined"

# --- LEXER ---
def tokenize(code):
    tokens = re.findall(r'\d+|[a-zA-Z_]\w*|==|<=|>=|[=+*-/<>]|{|}|\}|".*?"', code)
    return tokens

# --- AST NODES ---
class Number:
    def __init__(self, value): self.value = int(value)

class String:
    def __init__(self, value): self.value = value.strip('"')

class Var:
    def __init__(self, name): self.name = name

class BinOp:
    def __init__(self, left, op, right):
        self.left, self.op, self.right = left, op, right

# --- PARSER ---
def parse_expr(tokens):
    if not tokens:
        return None
    token = tokens.pop(0)
    if token.isdigit():
        left = Number(token)
    elif token.startswith('"') and token.endswith('"'):
        left = String(token)
    else:
        left = Var(token)

    if tokens and tokens[0] in ["+", "-", "*", "/"]:
        op = tokens.pop(0)
        right = parse_expr(tokens)
        return BinOp(left, op, right)

    return left

# --- INTERPRETER ---
def eval_ast(node, env, line_num=None, line=None):
    if isinstance(node, Number):
        return node.value
    if isinstance(node, String):
        return node.value
    if isinstance(node, Var):
        if node.name not in env:
            raise EngineNameError(node.name, line_num, line)
        return env[node.name]
    if isinstance(node, BinOp):
        left = eval_ast(node.left, env, line_num, line)
        right = eval_ast(node.right, env, line_num, line)
        if node.op == "+": return left + right
        if node.op == "-": return left - right
        if node.op == "*": return left * right
        if node.op == "/": return left // right
    return None

# --- MAIN RUNNER ---
def run(code, env=None):
    if env is None:
        env = {}
    lines = code.strip().split("\n")
    i = 0

    while i < len(lines):
        line = lines[i].strip()
        tokens = tokenize(line)

        if not tokens:
            i += 1
            continue

        if tokens[0] == "var":
            try:
                name = tokens[1]
                expr = parse_expr(tokens[3:])
                env[name] = eval_ast(expr, env, i+1, line)
            except Exception as e:
                print(e)
                return

        elif tokens[0] == "show":
            try:
                expr = parse_expr(tokens[1:])
                print(eval_ast(expr, env, i+1, line))
            except Exception as e:
                print(e)
                return

        elif tokens[0] == "if":
            condition_tokens = []
            for t in tokens[1:]:
                if t == "{":
                    break
                condition_tokens.append(t)
            condition = parse_expr(condition_tokens)
            cond_value = eval_ast(condition, env, i+1, line)

            # collect block
            block_lines = []
            i += 1
            while i < len(lines) and "}" not in lines[i]:
                block_lines.append(lines[i])
                i += 1

            if cond_value:
                run("\n".join(block_lines), env)

        elif tokens[0] == "FOREVER":
            block_lines = []
            i += 1
            while i < len(lines) and "}" not in lines[i]:
                block_lines.append(lines[i])
                i += 1

            while True:
                run("\n".join(block_lines), env)

        else:
            print(EngineSyntaxError(f"Unknown statement: {tokens[0]}", i+1, line))
            return

        i += 1

# --- FILE RUNNER ---
def run_file(filename):
    """Read and execute a .eng file."""
    try:
        with open(filename, 'r') as file:
            code = file.read()
        run(code)
    except FileNotFoundError:
        print(f"ENGINE-INTERPRETER: ERROR")
        print(f"File '{filename}' not found.")
    except Exception as e:
        print(f"ENGINE-INTERPRETER: ERROR")
        print(f"Error reading file '{filename}': {e}")

# --- COMMAND LINE ENTRY POINT ---
if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python -m engine <filename.eng>")
        print("")
        print("Run Engine language files (.eng) through the interpreter.")
        print("")
        print("Example:")
        print("  python -m engine my_program.eng")
        sys.exit(1)
    
    filename = sys.argv[1]
    if not filename.endswith('.eng'):
        print(f"Warning: '{filename}' does not have a .eng extension.")
    
    run_file(filename)
