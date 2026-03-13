"""
Engine Language Interpreter
===========================
A minimal, readable interpreted programming language.

Architecture: Source code -> Lexer (tokens) -> Parser (AST) -> Interpreter (evaluate)
"""

import re
import sys
import os

# ============================================================================
# ERRORS
# ============================================================================

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
        loc = f" (line {self.line_num})" if self.line_num else ""
        src = f"\n    {self.line}" if self.line else ""
        return f"[SyntaxError]{loc}: {self.message}{src}"


class EngineNameError(EngineError):
    def __init__(self, name, line_num=None, line=None):
        self.name = name
        self.line_num = line_num
        self.line = line
        super().__init__(self.__str__())

    def __str__(self):
        loc = f" (line {self.line_num})" if self.line_num else ""
        src = f"\n    {self.line}" if self.line else ""
        return f"[NameError]{loc}: Variable '{self.name}' is not defined.{src}\n  Hint: Did you forget to declare it with 'var'?"


class EngineTypeError(EngineError):
    def __init__(self, message, line_num=None, line=None):
        self.message = message
        self.line_num = line_num
        self.line = line
        super().__init__(self.__str__())

    def __str__(self):
        loc = f" (line {self.line_num})" if self.line_num else ""
        src = f"\n    {self.line}" if self.line else ""
        return f"[TypeError]{loc}: {self.message}{src}"


class EngineReturnSignal(Exception):
    """Internal signal used to propagate return values up the call stack."""
    def __init__(self, value):
        self.value = value


# ============================================================================
# TOKENS
# ============================================================================

class Token:
    __slots__ = ("type", "value", "line")

    def __init__(self, type_, value, line=0):
        self.type = type_
        self.value = value
        self.line = line

    def __repr__(self):
        return f"Token({self.type}, {self.value!r})"


# Token type constants
TT_INT       = "INT"
TT_FLOAT     = "FLOAT"
TT_STRING    = "STRING"
TT_IDENT     = "IDENT"
TT_KEYWORD   = "KEYWORD"
TT_PLUS      = "PLUS"
TT_MINUS     = "MINUS"
TT_MUL       = "MUL"
TT_DIV       = "DIV"
TT_MOD       = "MOD"
TT_EQ        = "EQ"
TT_EQEQ      = "EQEQ"
TT_NEQ       = "NEQ"
TT_LT        = "LT"
TT_GT        = "GT"
TT_LTE       = "LTE"
TT_GTE       = "GTE"
TT_AND       = "AND"
TT_OR        = "OR"
TT_NOT       = "NOT"
TT_LPAREN    = "LPAREN"
TT_RPAREN    = "RPAREN"
TT_LBRACE    = "LBRACE"
TT_RBRACE    = "RBRACE"
TT_LBRACKET  = "LBRACKET"
TT_RBRACKET  = "RBRACKET"
TT_COMMA     = "COMMA"
TT_DOT       = "DOT"
TT_NEWLINE   = "NEWLINE"
TT_EOF       = "EOF"

KEYWORDS = {
    "var", "show", "if", "else", "while", "for", "in",
    "func", "return", "true", "false", "null", "and", "or", "not",
    "import",
}


# ============================================================================
# LEXER
# ============================================================================

_TOKEN_SPEC = [
    ("COMMENT",   r"#[^\n]*"),
    ("FLOAT",     r"\d+\.\d+"),
    ("INT",       r"\d+"),
    ("STRING",    r'"(?:[^"\\]|\\.)*"|\'(?:[^\'\\]|\\.)*\''),
    ("IDENT",     r"[a-zA-Z_]\w*"),
    ("NEQ",       r"!="),
    ("EQEQ",     r"=="),
    ("LTE",       r"<="),
    ("GTE",       r">="),
    ("EQ",        r"="),
    ("PLUS",      r"\+"),
    ("MINUS",     r"-"),
    ("MUL",       r"\*"),
    ("DIV",       r"/"),
    ("MOD",       r"%"),
    ("LT",        r"<"),
    ("GT",        r">"),
    ("LPAREN",    r"\("),
    ("RPAREN",    r"\)"),
    ("LBRACE",    r"\{"),
    ("RBRACE",    r"\}"),
    ("LBRACKET",  r"\["),
    ("RBRACKET",  r"\]"),
    ("COMMA",     r","),
    ("DOT",       r"\."),
    ("NEWLINE",   r"\n"),
    ("SKIP",      r"[ \t\r]+"),
    ("MISMATCH",  r"."),
]

_TOKEN_RE = re.compile("|".join(f"(?P<{name}>{pattern})" for name, pattern in _TOKEN_SPEC))


def tokenize(source):
    """Convert source code string into a list of Token objects."""
    tokens = []
    line_num = 1
    for mo in _TOKEN_RE.finditer(source):
        kind = mo.lastgroup
        value = mo.group()
        if kind == "COMMENT" or kind == "SKIP":
            if kind == "SKIP" and "\n" in value:
                line_num += value.count("\n")
            continue
        if kind == "NEWLINE":
            tokens.append(Token(TT_NEWLINE, "\\n", line_num))
            line_num += 1
            continue
        if kind == "INT":
            tokens.append(Token(TT_INT, int(value), line_num))
        elif kind == "FLOAT":
            tokens.append(Token(TT_FLOAT, float(value), line_num))
        elif kind == "STRING":
            tokens.append(Token(TT_STRING, value[1:-1], line_num))
        elif kind == "IDENT":
            if value in KEYWORDS:
                if value == "and":
                    tokens.append(Token(TT_AND, value, line_num))
                elif value == "or":
                    tokens.append(Token(TT_OR, value, line_num))
                elif value == "not":
                    tokens.append(Token(TT_NOT, value, line_num))
                else:
                    tokens.append(Token(TT_KEYWORD, value, line_num))
            else:
                tokens.append(Token(TT_IDENT, value, line_num))
        elif kind == "MISMATCH":
            raise EngineSyntaxError(f"Unexpected character '{value}'", line_num)
        else:
            tt = kind  # PLUS, MINUS, etc. — kind name matches token type
            tokens.append(Token(tt, value, line_num))

    tokens.append(Token(TT_EOF, None, line_num))
    return tokens


# ============================================================================
# AST NODES
# ============================================================================

class NumberNode:
    __slots__ = ("value",)
    def __init__(self, value): self.value = value

class StringNode:
    __slots__ = ("value",)
    def __init__(self, value): self.value = value

class BoolNode:
    __slots__ = ("value",)
    def __init__(self, value): self.value = value

class NullNode:
    pass

class ListNode:
    __slots__ = ("elements",)
    def __init__(self, elements): self.elements = elements

class VarAccessNode:
    __slots__ = ("name", "line")
    def __init__(self, name, line=0):
        self.name = name
        self.line = line

class VarAssignNode:
    __slots__ = ("name", "value", "line")
    def __init__(self, name, value, line=0):
        self.name = name
        self.value = value
        self.line = line

class VarReassignNode:
    __slots__ = ("name", "value", "line")
    def __init__(self, name, value, line=0):
        self.name = name
        self.value = value
        self.line = line

class BinOpNode:
    __slots__ = ("left", "op", "right")
    def __init__(self, left, op, right):
        self.left = left
        self.op = op
        self.right = right

class UnaryOpNode:
    __slots__ = ("op", "operand")
    def __init__(self, op, operand):
        self.op = op
        self.operand = operand

class ShowNode:
    __slots__ = ("value",)
    def __init__(self, value): self.value = value

class IfNode:
    __slots__ = ("condition", "body", "else_body")
    def __init__(self, condition, body, else_body=None):
        self.condition = condition
        self.body = body
        self.else_body = else_body

class WhileNode:
    __slots__ = ("condition", "body")
    def __init__(self, condition, body):
        self.condition = condition
        self.body = body

class ForNode:
    __slots__ = ("var_name", "iterable", "body")
    def __init__(self, var_name, iterable, body):
        self.var_name = var_name
        self.iterable = iterable
        self.body = body

class FuncDefNode:
    __slots__ = ("name", "params", "body")
    def __init__(self, name, params, body):
        self.name = name
        self.params = params
        self.body = body

class FuncCallNode:
    __slots__ = ("name", "args", "line")
    def __init__(self, name, args, line=0):
        self.name = name
        self.args = args
        self.line = line

class ReturnNode:
    __slots__ = ("value",)
    def __init__(self, value): self.value = value

class IndexNode:
    __slots__ = ("obj", "index")
    def __init__(self, obj, index):
        self.obj = obj
        self.index = index

class MethodCallNode:
    __slots__ = ("obj", "method", "args", "line")
    def __init__(self, obj, method, args, line=0):
        self.obj = obj
        self.method = method
        self.args = args
        self.line = line

class ImportNode:
    __slots__ = ("path", "line")
    def __init__(self, path, line=0):
        self.path = path
        self.line = line

class BlockNode:
    __slots__ = ("statements",)
    def __init__(self, statements): self.statements = statements


# ============================================================================
# PARSER  (recursive descent with operator precedence)
# ============================================================================

class Parser:
    def __init__(self, tokens):
        self.tokens = tokens
        self.pos = 0

    # -- helpers --

    def current(self):
        return self.tokens[self.pos] if self.pos < len(self.tokens) else Token(TT_EOF, None)

    def peek(self, offset=1):
        idx = self.pos + offset
        return self.tokens[idx] if idx < len(self.tokens) else Token(TT_EOF, None)

    def eat(self, type_):
        tok = self.current()
        if tok.type != type_:
            raise EngineSyntaxError(
                f"Expected {type_}, got {tok.type} ('{tok.value}')", tok.line
            )
        self.pos += 1
        return tok

    def skip_newlines(self):
        while self.current().type == TT_NEWLINE:
            self.pos += 1

    # -- entry --

    def parse(self):
        """Parse the entire token stream into a BlockNode."""
        statements = []
        self.skip_newlines()
        while self.current().type != TT_EOF:
            stmt = self.statement()
            if stmt is not None:
                statements.append(stmt)
            self.skip_newlines()
        return BlockNode(statements)

    # -- statements --

    def statement(self):
        tok = self.current()

        if tok.type == TT_KEYWORD:
            if tok.value == "var":
                return self.var_assign()
            if tok.value == "show":
                return self.show_stmt()
            if tok.value == "if":
                return self.if_stmt()
            if tok.value == "while":
                return self.while_stmt()
            if tok.value == "for":
                return self.for_stmt()
            if tok.value == "func":
                return self.func_def()
            if tok.value == "return":
                return self.return_stmt()
            if tok.value == "import":
                return self.import_stmt()

        # Bare identifier: could be reassignment (x = ...) or expression statement (fn call)
        if tok.type == TT_IDENT and self.peek().type == TT_EQ:
            return self.var_reassign()

        # Expression statement (e.g. a function call on its own line)
        expr = self.expr()
        return expr

    def var_assign(self):
        self.eat(TT_KEYWORD)  # var
        name_tok = self.eat(TT_IDENT)
        self.eat(TT_EQ)
        value = self.expr()
        return VarAssignNode(name_tok.value, value, name_tok.line)

    def var_reassign(self):
        name_tok = self.eat(TT_IDENT)
        self.eat(TT_EQ)
        value = self.expr()
        return VarReassignNode(name_tok.value, value, name_tok.line)

    def show_stmt(self):
        self.eat(TT_KEYWORD)  # show
        value = self.expr()
        return ShowNode(value)

    def if_stmt(self):
        self.eat(TT_KEYWORD)  # if
        condition = self.expr()
        body = self.block()
        else_body = None
        self.skip_newlines()
        if self.current().type == TT_KEYWORD and self.current().value == "else":
            self.eat(TT_KEYWORD)  # else
            if self.current().type == TT_KEYWORD and self.current().value == "if":
                else_body = [self.if_stmt()]
            else:
                else_body = self.block()
        return IfNode(condition, body, else_body)

    def while_stmt(self):
        self.eat(TT_KEYWORD)  # while
        condition = self.expr()
        body = self.block()
        return WhileNode(condition, body)

    def for_stmt(self):
        self.eat(TT_KEYWORD)  # for
        var_tok = self.eat(TT_IDENT)
        self.eat(TT_KEYWORD)  # in
        iterable = self.expr()
        body = self.block()
        return ForNode(var_tok.value, iterable, body)

    def func_def(self):
        self.eat(TT_KEYWORD)  # func
        name_tok = self.eat(TT_IDENT)
        self.eat(TT_LPAREN)
        params = []
        if self.current().type != TT_RPAREN:
            params.append(self.eat(TT_IDENT).value)
            while self.current().type == TT_COMMA:
                self.eat(TT_COMMA)
                params.append(self.eat(TT_IDENT).value)
        self.eat(TT_RPAREN)
        body = self.block()
        return FuncDefNode(name_tok.value, params, body)

    def return_stmt(self):
        self.eat(TT_KEYWORD)  # return
        value = None
        if self.current().type not in (TT_NEWLINE, TT_EOF, TT_RBRACE):
            value = self.expr()
        return ReturnNode(value)

    def import_stmt(self):
        tok = self.eat(TT_KEYWORD)  # import
        path_tok = self.eat(TT_STRING)
        return ImportNode(path_tok.value, tok.line)

    def block(self):
        """Parse a { ... } block, returning a list of statement nodes."""
        self.skip_newlines()
        self.eat(TT_LBRACE)
        self.skip_newlines()
        stmts = []
        while self.current().type != TT_RBRACE:
            if self.current().type == TT_EOF:
                raise EngineSyntaxError("Expected '}' but reached end of file", self.current().line)
            stmt = self.statement()
            if stmt is not None:
                stmts.append(stmt)
            self.skip_newlines()
        self.eat(TT_RBRACE)
        return stmts

    # -- expressions (precedence climbing) --

    def expr(self):
        return self.or_expr()

    def or_expr(self):
        left = self.and_expr()
        while self.current().type == TT_OR:
            op = self.eat(TT_OR).value
            right = self.and_expr()
            left = BinOpNode(left, op, right)
        return left

    def and_expr(self):
        left = self.not_expr()
        while self.current().type == TT_AND:
            op = self.eat(TT_AND).value
            right = self.not_expr()
            left = BinOpNode(left, op, right)
        return left

    def not_expr(self):
        if self.current().type == TT_NOT:
            op = self.eat(TT_NOT).value
            operand = self.not_expr()
            return UnaryOpNode(op, operand)
        return self.comparison()

    def comparison(self):
        left = self.arith_expr()
        while self.current().type in (TT_EQEQ, TT_NEQ, TT_LT, TT_GT, TT_LTE, TT_GTE):
            op = self.current().value
            self.pos += 1
            right = self.arith_expr()
            left = BinOpNode(left, op, right)
        return left

    def arith_expr(self):
        left = self.term()
        while self.current().type in (TT_PLUS, TT_MINUS):
            op = self.current().value
            self.pos += 1
            right = self.term()
            left = BinOpNode(left, op, right)
        return left

    def term(self):
        left = self.unary()
        while self.current().type in (TT_MUL, TT_DIV, TT_MOD):
            op = self.current().value
            self.pos += 1
            right = self.unary()
            left = BinOpNode(left, op, right)
        return left

    def unary(self):
        if self.current().type == TT_MINUS:
            self.pos += 1
            operand = self.unary()
            return UnaryOpNode("-", operand)
        return self.postfix()

    def postfix(self):
        """Handle call (), index [], and dot . access after an atom."""
        node = self.atom()
        while True:
            if self.current().type == TT_LPAREN:
                # function call
                if isinstance(node, VarAccessNode):
                    name = node.name
                    line = node.line
                else:
                    raise EngineSyntaxError("Can only call functions by name", self.current().line)
                self.eat(TT_LPAREN)
                args = []
                if self.current().type != TT_RPAREN:
                    args.append(self.expr())
                    while self.current().type == TT_COMMA:
                        self.eat(TT_COMMA)
                        args.append(self.expr())
                self.eat(TT_RPAREN)
                node = FuncCallNode(name, args, line)
            elif self.current().type == TT_LBRACKET:
                self.eat(TT_LBRACKET)
                index = self.expr()
                self.eat(TT_RBRACKET)
                node = IndexNode(node, index)
            elif self.current().type == TT_DOT:
                self.eat(TT_DOT)
                method_tok = self.eat(TT_IDENT)
                if self.current().type == TT_LPAREN:
                    self.eat(TT_LPAREN)
                    args = []
                    if self.current().type != TT_RPAREN:
                        args.append(self.expr())
                        while self.current().type == TT_COMMA:
                            self.eat(TT_COMMA)
                            args.append(self.expr())
                    self.eat(TT_RPAREN)
                    node = MethodCallNode(node, method_tok.value, args, method_tok.line)
                else:
                    # property access — treat as method call with no args for now
                    node = MethodCallNode(node, method_tok.value, [], method_tok.line)
            else:
                break
        return node

    def atom(self):
        tok = self.current()

        if tok.type == TT_INT:
            self.pos += 1
            return NumberNode(tok.value)

        if tok.type == TT_FLOAT:
            self.pos += 1
            return NumberNode(tok.value)

        if tok.type == TT_STRING:
            self.pos += 1
            return StringNode(tok.value)

        if tok.type == TT_KEYWORD and tok.value == "true":
            self.pos += 1
            return BoolNode(True)

        if tok.type == TT_KEYWORD and tok.value == "false":
            self.pos += 1
            return BoolNode(False)

        if tok.type == TT_KEYWORD and tok.value == "null":
            self.pos += 1
            return NullNode()

        if tok.type == TT_IDENT:
            self.pos += 1
            return VarAccessNode(tok.value, tok.line)

        if tok.type == TT_LPAREN:
            self.eat(TT_LPAREN)
            expr = self.expr()
            self.eat(TT_RPAREN)
            return expr

        if tok.type == TT_LBRACKET:
            return self.list_literal()

        raise EngineSyntaxError(
            f"Unexpected token '{tok.value}'", tok.line
        )

    def list_literal(self):
        self.eat(TT_LBRACKET)
        elements = []
        self.skip_newlines()
        if self.current().type != TT_RBRACKET:
            elements.append(self.expr())
            while self.current().type == TT_COMMA:
                self.eat(TT_COMMA)
                self.skip_newlines()
                elements.append(self.expr())
        self.skip_newlines()
        self.eat(TT_RBRACKET)
        return ListNode(elements)


# ============================================================================
# ENVIRONMENT
# ============================================================================

class Environment:
    """Variable scope — supports nested (local) scopes."""
    def __init__(self, parent=None):
        self.store = {}
        self.parent = parent

    def get(self, name):
        if name in self.store:
            return self.store[name]
        if self.parent:
            return self.parent.get(name)
        return None

    def has(self, name):
        if name in self.store:
            return True
        if self.parent:
            return self.parent.has(name)
        return False

    def set(self, name, value):
        self.store[name] = value

    def set_existing(self, name, value):
        """Reassign a variable that already exists (walk up scopes)."""
        if name in self.store:
            self.store[name] = value
            return True
        if self.parent:
            return self.parent.set_existing(name, value)
        return False


# ============================================================================
# ENGINE FUNCTION wrapper
# ============================================================================

class EngineFunction:
    """Stores a user-defined function."""
    __slots__ = ("name", "params", "body", "closure_env")
    def __init__(self, name, params, body, closure_env):
        self.name = name
        self.params = params
        self.body = body
        self.closure_env = closure_env


# ============================================================================
# INTERPRETER
# ============================================================================

class Interpreter:
    def __init__(self, env=None, source_dir=None):
        self.global_env = env if env else Environment()
        self.source_dir = source_dir or os.getcwd()
        self._register_builtins()

    # -- built-in functions --

    def _register_builtins(self):
        self.builtins = {
            "show":    self._builtin_show,
            "len":     self._builtin_len,
            "type":    self._builtin_type,
            "int":     self._builtin_int,
            "float":   self._builtin_float,
            "str":     self._builtin_str,
            "input":   self._builtin_input,
            "range":   self._builtin_range,
            "push":    self._builtin_push,
            "pop":     self._builtin_pop,
            "abs":     self._builtin_abs,
            "min":     self._builtin_min,
            "max":     self._builtin_max,
            "round":   self._builtin_round,
        }

    @staticmethod
    def _builtin_show(*args):
        print(*args)
        return None

    @staticmethod
    def _builtin_len(obj):
        return len(obj)

    @staticmethod
    def _builtin_type(obj):
        if obj is None:            return "null"
        if isinstance(obj, bool):  return "bool"
        if isinstance(obj, int):   return "int"
        if isinstance(obj, float): return "float"
        if isinstance(obj, str):   return "string"
        if isinstance(obj, list):  return "list"
        return "unknown"

    @staticmethod
    def _builtin_int(val):
        return int(val)

    @staticmethod
    def _builtin_float(val):
        return float(val)

    @staticmethod
    def _builtin_str(val):
        if val is None: return "null"
        if isinstance(val, bool): return "true" if val else "false"
        return str(val)

    @staticmethod
    def _builtin_input(prompt=""):
        return input(prompt)

    @staticmethod
    def _builtin_range(*args):
        return list(range(*args))

    @staticmethod
    def _builtin_push(lst, item):
        lst.append(item)
        return lst

    @staticmethod
    def _builtin_pop(lst):
        return lst.pop()

    @staticmethod
    def _builtin_abs(val):
        return abs(val)

    @staticmethod
    def _builtin_min(*args):
        if len(args) == 1 and isinstance(args[0], list):
            return min(args[0])
        return min(args)

    @staticmethod
    def _builtin_max(*args):
        if len(args) == 1 and isinstance(args[0], list):
            return max(args[0])
        return max(args)

    @staticmethod
    def _builtin_round(val, ndigits=0):
        return round(val, ndigits)

    # -- method dispatch on values --

    def _call_method(self, obj, method, args, line=0):
        # String methods
        if isinstance(obj, str):
            if method == "length":  return len(obj)
            if method == "upper":   return obj.upper()
            if method == "lower":   return obj.lower()
            if method == "split":   return obj.split(args[0]) if args else obj.split()
            if method == "replace": return obj.replace(args[0], args[1])
            if method == "contains":return args[0] in obj
            if method == "starts_with": return obj.startswith(args[0])
            if method == "ends_with":   return obj.endswith(args[0])
            if method == "strip":   return obj.strip()
            if method == "find":    return obj.find(args[0])
            raise EngineTypeError(f"String has no method '{method}'", line)

        # List methods
        if isinstance(obj, list):
            if method == "length":  return len(obj)
            if method == "push":    obj.append(args[0]); return obj
            if method == "pop":     return obj.pop()
            if method == "join":    return args[0].join(str(x) for x in obj)
            if method == "contains":return args[0] in obj
            if method == "reverse": obj.reverse(); return obj
            if method == "sort":    obj.sort(); return obj
            if method == "slice":
                start = args[0] if len(args) > 0 else 0
                end   = args[1] if len(args) > 1 else len(obj)
                return obj[start:end]
            raise EngineTypeError(f"List has no method '{method}'", line)

        raise EngineTypeError(f"Cannot call method '{method}' on {type(obj).__name__}", line)

    # -- core evaluate --

    def visit(self, node, env):
        """Evaluate a single AST node."""
        method_name = f"visit_{type(node).__name__}"
        visitor = getattr(self, method_name, None)
        if visitor is None:
            raise EngineError(f"No visit method for {type(node).__name__}")
        return visitor(node, env)

    def visit_BlockNode(self, node, env):
        result = None
        for stmt in node.statements:
            result = self.visit(stmt, env)
        return result

    def visit_NumberNode(self, node, env):
        return node.value

    def visit_StringNode(self, node, env):
        return node.value

    def visit_BoolNode(self, node, env):
        return node.value

    def visit_NullNode(self, node, env):
        return None

    def visit_ListNode(self, node, env):
        return [self.visit(el, env) for el in node.elements]

    def visit_VarAccessNode(self, node, env):
        if env.has(node.name):
            return env.get(node.name)
        raise EngineNameError(node.name, node.line)

    def visit_VarAssignNode(self, node, env):
        value = self.visit(node.value, env)
        env.set(node.name, value)
        return value

    def visit_VarReassignNode(self, node, env):
        value = self.visit(node.value, env)
        if not env.set_existing(node.name, value):
            raise EngineNameError(node.name, node.line)
        return value

    def visit_BinOpNode(self, node, env):
        # short-circuit for logical operators
        if node.op == "and":
            left = self.visit(node.left, env)
            if not left:
                return left
            return self.visit(node.right, env)
        if node.op == "or":
            left = self.visit(node.left, env)
            if left:
                return left
            return self.visit(node.right, env)

        left = self.visit(node.left, env)
        right = self.visit(node.right, env)

        if node.op == "+":
            if isinstance(left, str) or isinstance(right, str):
                return str(left if left is not None else "null") + str(right if right is not None else "null")
            return left + right
        if node.op == "-":  return left - right
        if node.op == "*":
            if isinstance(left, str) and isinstance(right, int):
                return left * right
            if isinstance(left, int) and isinstance(right, str):
                return right * left
            return left * right
        if node.op == "/":
            if right == 0:
                raise EngineTypeError("Division by zero")
            if isinstance(left, int) and isinstance(right, int):
                return left // right
            return left / right
        if node.op == "%":  return left % right
        if node.op == "==": return left == right
        if node.op == "!=": return left != right
        if node.op == "<":  return left < right
        if node.op == ">":  return left > right
        if node.op == "<=": return left <= right
        if node.op == ">=": return left >= right

        raise EngineSyntaxError(f"Unknown operator '{node.op}'")

    def visit_UnaryOpNode(self, node, env):
        val = self.visit(node.operand, env)
        if node.op == "-":
            return -val
        if node.op == "not":
            return not val
        raise EngineSyntaxError(f"Unknown unary operator '{node.op}'")

    def visit_ShowNode(self, node, env):
        value = self.visit(node.value, env)
        if value is None:
            print("null")
        elif isinstance(value, bool):
            print("true" if value else "false")
        else:
            print(value)
        return None

    def visit_IfNode(self, node, env):
        condition = self.visit(node.condition, env)
        if condition:
            for stmt in node.body:
                self.visit(stmt, env)
        elif node.else_body:
            for stmt in node.else_body:
                self.visit(stmt, env)
        return None

    def visit_WhileNode(self, node, env):
        while self.visit(node.condition, env):
            for stmt in node.body:
                self.visit(stmt, env)
        return None

    def visit_ForNode(self, node, env):
        iterable = self.visit(node.iterable, env)
        if not hasattr(iterable, "__iter__"):
            raise EngineTypeError(f"Cannot iterate over {type(iterable).__name__}")
        for item in iterable:
            env.set(node.var_name, item)
            for stmt in node.body:
                self.visit(stmt, env)
        return None

    def visit_FuncDefNode(self, node, env):
        func = EngineFunction(node.name, node.params, node.body, env)
        env.set(node.name, func)
        return func

    def visit_FuncCallNode(self, node, env):
        args = [self.visit(a, env) for a in node.args]

        # check builtins first
        if node.name in self.builtins:
            try:
                return self.builtins[node.name](*args)
            except TypeError as e:
                raise EngineTypeError(str(e), node.line)

        # user-defined function
        if not env.has(node.name):
            raise EngineNameError(node.name, node.line)

        func = env.get(node.name)
        if not isinstance(func, EngineFunction):
            raise EngineTypeError(f"'{node.name}' is not a function", node.line)

        if len(args) != len(func.params):
            raise EngineTypeError(
                f"Function '{func.name}' expects {len(func.params)} argument(s), got {len(args)}",
                node.line,
            )

        local_env = Environment(parent=func.closure_env)
        for pname, pval in zip(func.params, args):
            local_env.set(pname, pval)

        try:
            for stmt in func.body:
                self.visit(stmt, local_env)
        except EngineReturnSignal as ret:
            return ret.value

        return None

    def visit_ReturnNode(self, node, env):
        value = self.visit(node.value, env) if node.value else None
        raise EngineReturnSignal(value)

    def visit_IndexNode(self, node, env):
        obj = self.visit(node.obj, env)
        index = self.visit(node.index, env)
        try:
            return obj[index]
        except (IndexError, KeyError):
            raise EngineTypeError(f"Index {index} out of range")

    def visit_MethodCallNode(self, node, env):
        obj = self.visit(node.obj, env)
        args = [self.visit(a, env) for a in node.args]
        return self._call_method(obj, node.method, args, node.line)

    def visit_ImportNode(self, node, env):
        path = node.path
        if not path.endswith(".eng"):
            path += ".eng"
        full_path = os.path.join(self.source_dir, path)
        if not os.path.exists(full_path):
            raise EngineSyntaxError(f"Cannot find module '{path}'", node.line)
        with open(full_path, "r") as f:
            source = f.read()
        tokens = tokenize(source)
        parser = Parser(tokens)
        tree = parser.parse()
        self.visit(tree, env)
        return None


# ============================================================================
# PUBLIC API
# ============================================================================

def run(code, env=None):
    """Run Engine source code from a string."""
    try:
        tokens = tokenize(code)
        parser = Parser(tokens)
        tree = parser.parse()
        interp = Interpreter(env=Environment() if env is None else env)
        interp.visit(tree, interp.global_env)
    except EngineError as e:
        print(f"Engine Error: {e}")
        sys.exit(1)


def run_file(filename):
    """Read and execute a .eng file."""
    if not os.path.exists(filename):
        print(f"Engine Error: File '{filename}' not found.")
        sys.exit(1)
    try:
        with open(filename, "r") as f:
            source = f.read()
        source_dir = os.path.dirname(os.path.abspath(filename))
        tokens = tokenize(source)
        parser = Parser(tokens)
        tree = parser.parse()
        interp = Interpreter(source_dir=source_dir)
        interp.visit(tree, interp.global_env)
    except EngineError as e:
        print(f"Engine Error: {e}")
        sys.exit(1)
