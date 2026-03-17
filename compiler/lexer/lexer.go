package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	// Literals
	TOKEN_INT TokenType = iota
	TOKEN_FLOAT
	TOKEN_STRING
	TOKEN_IDENT
	TOKEN_KEYWORD
	// Operators
	TOKEN_PLUS
	TOKEN_MINUS
	TOKEN_STAR
	TOKEN_SLASH
	TOKEN_MOD
	TOKEN_EQ
	TOKEN_EQEQ
	TOKEN_NEQ
	TOKEN_LT
	TOKEN_GT
	TOKEN_LTE
	TOKEN_GTE
	TOKEN_AND
	TOKEN_OR
	TOKEN_NOT
	// Delimiters
	TOKEN_LPAREN
	TOKEN_RPAREN
	TOKEN_LBRACE
	TOKEN_RBRACE
	TOKEN_LBRACKET
	TOKEN_RBRACKET
	TOKEN_COMMA
	TOKEN_DOT
	TOKEN_NEWLINE
	TOKEN_EOF
)

var tokenNames = map[TokenType]string{
	TOKEN_INT: "INT", TOKEN_FLOAT: "FLOAT", TOKEN_STRING: "STRING",
	TOKEN_IDENT: "IDENT", TOKEN_KEYWORD: "KEYWORD",
	TOKEN_PLUS: "PLUS", TOKEN_MINUS: "MINUS", TOKEN_STAR: "STAR",
	TOKEN_SLASH: "SLASH", TOKEN_MOD: "MOD",
	TOKEN_EQ: "EQ", TOKEN_EQEQ: "EQEQ", TOKEN_NEQ: "NEQ",
	TOKEN_LT: "LT", TOKEN_GT: "GT", TOKEN_LTE: "LTE", TOKEN_GTE: "GTE",
	TOKEN_AND: "AND", TOKEN_OR: "OR", TOKEN_NOT: "NOT",
	TOKEN_LPAREN: "LPAREN", TOKEN_RPAREN: "RPAREN",
	TOKEN_LBRACE: "LBRACE", TOKEN_RBRACE: "RBRACE",
	TOKEN_LBRACKET: "LBRACKET", TOKEN_RBRACKET: "RBRACKET",
	TOKEN_COMMA: "COMMA", TOKEN_DOT: "DOT",
	TOKEN_NEWLINE: "NEWLINE", TOKEN_EOF: "EOF",
}

func (t TokenType) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN(%d)", int(t))
}

type Token struct {
	Type    TokenType
	Value   string
	Line    int
	Col     int
}

func (t Token) String() string {
	return fmt.Sprintf("Token(%s, %q, line=%d)", t.Type, t.Value, t.Line)
}

var keywords = map[string]bool{
	"var": true, "show": true, "if": true, "else": true,
	"while": true, "for": true, "in": true,
	"func": true, "return": true, "true": true, "false": true,
	"null": true, "and": true, "or": true, "not": true,
	"import": true,
}

func IsKeyword(s string) bool {
	return keywords[s]
}

type Lexer struct {
	source  []rune
	pos     int
	line    int
	col     int
	tokens  []Token
}

func New(source string) *Lexer {
	return &Lexer{
		source: []rune(source),
		pos:    0,
		line:   1,
		col:    1,
	}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.source) {
		return 0
	}
	return l.source[l.pos]
}

func (l *Lexer) advance() rune {
	ch := l.source[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) addToken(tt TokenType, value string, line, col int) {
	l.tokens = append(l.tokens, Token{Type: tt, Value: value, Line: line, Col: col})
}

func (l *Lexer) Tokenize() ([]Token, error) {
	for l.pos < len(l.source) {
		ch := l.peek()

		// Skip whitespace (not newlines)
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.advance()
			continue
		}

		// Newline
		if ch == '\n' {
			line, col := l.line, l.col
			l.advance()
			l.addToken(TOKEN_NEWLINE, "\\n", line, col)
			continue
		}

		// Comments
		if ch == '#' {
			for l.pos < len(l.source) && l.peek() != '\n' {
				l.advance()
			}
			continue
		}

		// Numbers
		if unicode.IsDigit(ch) {
			l.readNumber()
			continue
		}

		// Strings
		if ch == '"' || ch == '\'' {
			err := l.readString()
			if err != nil {
				return nil, err
			}
			continue
		}

		// Identifiers and keywords
		if unicode.IsLetter(ch) || ch == '_' {
			l.readIdent()
			continue
		}

		// Two-character operators
		line, col := l.line, l.col
		if l.pos+1 < len(l.source) {
			two := string(l.source[l.pos : l.pos+2])
			switch two {
			case "==":
				l.advance(); l.advance()
				l.addToken(TOKEN_EQEQ, "==", line, col)
				continue
			case "!=":
				l.advance(); l.advance()
				l.addToken(TOKEN_NEQ, "!=", line, col)
				continue
			case "<=":
				l.advance(); l.advance()
				l.addToken(TOKEN_LTE, "<=", line, col)
				continue
			case ">=":
				l.advance(); l.advance()
				l.addToken(TOKEN_GTE, ">=", line, col)
				continue
			}
		}

		// Single-character operators
		l.advance()
		switch ch {
		case '+':
			l.addToken(TOKEN_PLUS, "+", line, col)
		case '-':
			l.addToken(TOKEN_MINUS, "-", line, col)
		case '*':
			l.addToken(TOKEN_STAR, "*", line, col)
		case '/':
			l.addToken(TOKEN_SLASH, "/", line, col)
		case '%':
			l.addToken(TOKEN_MOD, "%", line, col)
		case '=':
			l.addToken(TOKEN_EQ, "=", line, col)
		case '<':
			l.addToken(TOKEN_LT, "<", line, col)
		case '>':
			l.addToken(TOKEN_GT, ">", line, col)
		case '(':
			l.addToken(TOKEN_LPAREN, "(", line, col)
		case ')':
			l.addToken(TOKEN_RPAREN, ")", line, col)
		case '{':
			l.addToken(TOKEN_LBRACE, "{", line, col)
		case '}':
			l.addToken(TOKEN_RBRACE, "}", line, col)
		case '[':
			l.addToken(TOKEN_LBRACKET, "[", line, col)
		case ']':
			l.addToken(TOKEN_RBRACKET, "]", line, col)
		case ',':
			l.addToken(TOKEN_COMMA, ",", line, col)
		case '.':
			l.addToken(TOKEN_DOT, ".", line, col)
		default:
			return nil, fmt.Errorf("[SyntaxError] (line %d): Unexpected character '%c'", line, ch)
		}
	}

	l.addToken(TOKEN_EOF, "", l.line, l.col)
	return l.tokens, nil
}

func (l *Lexer) readNumber() {
	line, col := l.line, l.col
	var sb strings.Builder
	isFloat := false

	for l.pos < len(l.source) && (unicode.IsDigit(l.peek()) || l.peek() == '.') {
		if l.peek() == '.' {
			if isFloat {
				break
			}
			isFloat = true
		}
		sb.WriteRune(l.advance())
	}

	if isFloat {
		l.addToken(TOKEN_FLOAT, sb.String(), line, col)
	} else {
		l.addToken(TOKEN_INT, sb.String(), line, col)
	}
}

func (l *Lexer) readString() error {
	line, col := l.line, l.col
	quote := l.advance() // consume opening quote
	var sb strings.Builder

	for l.pos < len(l.source) && l.peek() != quote {
		ch := l.peek()
		if ch == '\\' {
			l.advance()
			if l.pos >= len(l.source) {
				return fmt.Errorf("[SyntaxError] (line %d): Unterminated string escape", line)
			}
			esc := l.advance()
			switch esc {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case '\\':
				sb.WriteRune('\\')
			case '\'':
				sb.WriteRune('\'')
			case '"':
				sb.WriteRune('"')
			default:
				sb.WriteRune('\\')
				sb.WriteRune(esc)
			}
		} else {
			sb.WriteRune(l.advance())
		}
	}

	if l.pos >= len(l.source) {
		return fmt.Errorf("[SyntaxError] (line %d): Unterminated string", line)
	}
	l.advance() // consume closing quote
	l.addToken(TOKEN_STRING, sb.String(), line, col)
	return nil
}

func (l *Lexer) readIdent() {
	line, col := l.line, l.col
	var sb strings.Builder

	for l.pos < len(l.source) && (unicode.IsLetter(l.peek()) || unicode.IsDigit(l.peek()) || l.peek() == '_') {
		sb.WriteRune(l.advance())
	}

	word := sb.String()
	switch word {
	case "and":
		l.addToken(TOKEN_AND, word, line, col)
	case "or":
		l.addToken(TOKEN_OR, word, line, col)
	case "not":
		l.addToken(TOKEN_NOT, word, line, col)
	default:
		if IsKeyword(word) {
			l.addToken(TOKEN_KEYWORD, word, line, col)
		} else {
			l.addToken(TOKEN_IDENT, word, line, col)
		}
	}
}
