package parser

import (
	"fmt"

	"github.com/annuaicoder/yoft/compiler/ast"
	"github.com/annuaicoder/yoft/compiler/lexer"
)

type Parser struct {
	tokens []lexer.Token
	pos    int
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) current() lexer.Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return lexer.Token{Type: lexer.TOKEN_EOF}
}

func (p *Parser) peek(offset int) lexer.Token {
	idx := p.pos + offset
	if idx < len(p.tokens) {
		return p.tokens[idx]
	}
	return lexer.Token{Type: lexer.TOKEN_EOF}
}

func (p *Parser) eat(tt lexer.TokenType) (lexer.Token, error) {
	tok := p.current()
	if tok.Type != tt {
		return tok, fmt.Errorf("[SyntaxError] (line %d): Expected %s, got %s ('%s')", tok.Line, tt, tok.Type, tok.Value)
	}
	p.pos++
	return tok, nil
}

func (p *Parser) skipNewlines() {
	for p.current().Type == lexer.TOKEN_NEWLINE {
		p.pos++
	}
}

func (p *Parser) Parse() (*ast.Program, error) {
	stmts := []ast.Node{}
	p.skipNewlines()
	for p.current().Type != lexer.TOKEN_EOF {
		stmt, err := p.statement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
		p.skipNewlines()
	}
	return &ast.Program{Statements: stmts}, nil
}

func (p *Parser) statement() (ast.Node, error) {
	tok := p.current()
	if tok.Type == lexer.TOKEN_KEYWORD {
		switch tok.Value {
		case "var":
			return p.varDecl()
		case "show":
			return p.showStmt()
		case "if":
			return p.ifStmt()
		case "while":
			return p.whileStmt()
		case "for":
			return p.forStmt()
		case "func":
			return p.funcDecl()
		case "return":
			return p.returnStmt()
		case "import":
			return p.importStmt()
		}
	}
	if tok.Type == lexer.TOKEN_IDENT && p.peek(1).Type == lexer.TOKEN_EQ {
		return p.varReassign()
	}
	return p.expr()
}

func (p *Parser) varDecl() (ast.Node, error) {
	p.eat(lexer.TOKEN_KEYWORD)
	nameTok, err := p.eat(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}
	if _, err := p.eat(lexer.TOKEN_EQ); err != nil {
		return nil, err
	}
	value, err := p.expr()
	if err != nil {
		return nil, err
	}
	return &ast.VarDecl{Name: nameTok.Value, Value: value, Line: nameTok.Line}, nil
}

func (p *Parser) varReassign() (ast.Node, error) {
	nameTok, _ := p.eat(lexer.TOKEN_IDENT)
	p.eat(lexer.TOKEN_EQ)
	value, err := p.expr()
	if err != nil {
		return nil, err
	}
	return &ast.VarReassign{Name: nameTok.Value, Value: value, Line: nameTok.Line}, nil
}

func (p *Parser) showStmt() (ast.Node, error) {
	tok, _ := p.eat(lexer.TOKEN_KEYWORD)
	value, err := p.expr()
	if err != nil {
		return nil, err
	}
	return &ast.ShowStmt{Value: value, Line: tok.Line}, nil
}

func (p *Parser) ifStmt() (ast.Node, error) {
	tok, _ := p.eat(lexer.TOKEN_KEYWORD)
	cond, err := p.expr()
	if err != nil {
		return nil, err
	}
	body, err := p.block()
	if err != nil {
		return nil, err
	}
	var elseBody []ast.Node
	p.skipNewlines()
	if p.current().Type == lexer.TOKEN_KEYWORD && p.current().Value == "else" {
		p.eat(lexer.TOKEN_KEYWORD)
		if p.current().Type == lexer.TOKEN_KEYWORD && p.current().Value == "if" {
			elseIf, err := p.ifStmt()
			if err != nil {
				return nil, err
			}
			elseBody = []ast.Node{elseIf}
		} else {
			elseBody, err = p.block()
			if err != nil {
				return nil, err
			}
		}
	}
	return &ast.IfStmt{Condition: cond, Body: body, ElseBody: elseBody, Line: tok.Line}, nil
}

func (p *Parser) whileStmt() (ast.Node, error) {
	tok, _ := p.eat(lexer.TOKEN_KEYWORD)
	cond, err := p.expr()
	if err != nil {
		return nil, err
	}
	body, err := p.block()
	if err != nil {
		return nil, err
	}
	return &ast.WhileStmt{Condition: cond, Body: body, Line: tok.Line}, nil
}

func (p *Parser) forStmt() (ast.Node, error) {
	tok, _ := p.eat(lexer.TOKEN_KEYWORD)
	varTok, err := p.eat(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}
	p.eat(lexer.TOKEN_KEYWORD) // in
	iter, err := p.expr()
	if err != nil {
		return nil, err
	}
	body, err := p.block()
	if err != nil {
		return nil, err
	}
	return &ast.ForStmt{VarName: varTok.Value, Iterable: iter, Body: body, Line: tok.Line}, nil
}

func (p *Parser) funcDecl() (ast.Node, error) {
	tok, _ := p.eat(lexer.TOKEN_KEYWORD)
	nameTok, err := p.eat(lexer.TOKEN_IDENT)
	if err != nil {
		return nil, err
	}
	p.eat(lexer.TOKEN_LPAREN)
	params := []string{}
	if p.current().Type != lexer.TOKEN_RPAREN {
		pt, err := p.eat(lexer.TOKEN_IDENT)
		if err != nil {
			return nil, err
		}
		params = append(params, pt.Value)
		for p.current().Type == lexer.TOKEN_COMMA {
			p.eat(lexer.TOKEN_COMMA)
			pt, err := p.eat(lexer.TOKEN_IDENT)
			if err != nil {
				return nil, err
			}
			params = append(params, pt.Value)
		}
	}
	p.eat(lexer.TOKEN_RPAREN)
	body, err := p.block()
	if err != nil {
		return nil, err
	}
	return &ast.FuncDecl{Name: nameTok.Value, Params: params, Body: body, Line: tok.Line}, nil
}

func (p *Parser) returnStmt() (ast.Node, error) {
	tok, _ := p.eat(lexer.TOKEN_KEYWORD)
	var value ast.Node
	cur := p.current()
	if cur.Type != lexer.TOKEN_NEWLINE && cur.Type != lexer.TOKEN_EOF && cur.Type != lexer.TOKEN_RBRACE {
		var err error
		value, err = p.expr()
		if err != nil {
			return nil, err
		}
	}
	return &ast.ReturnStmt{Value: value, Line: tok.Line}, nil
}

func (p *Parser) importStmt() (ast.Node, error) {
	tok, _ := p.eat(lexer.TOKEN_KEYWORD)
	pathTok, err := p.eat(lexer.TOKEN_STRING)
	if err != nil {
		return nil, err
	}
	return &ast.ImportStmt{Path: pathTok.Value, Line: tok.Line}, nil
}

func (p *Parser) block() ([]ast.Node, error) {
	p.skipNewlines()
	if _, err := p.eat(lexer.TOKEN_LBRACE); err != nil {
		return nil, err
	}
	p.skipNewlines()
	stmts := []ast.Node{}
	for p.current().Type != lexer.TOKEN_RBRACE {
		if p.current().Type == lexer.TOKEN_EOF {
			return nil, fmt.Errorf("[SyntaxError] (line %d): Expected '}' but reached end of file", p.current().Line)
		}
		stmt, err := p.statement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
		p.skipNewlines()
	}
	p.eat(lexer.TOKEN_RBRACE)
	return stmts, nil
}

func (p *Parser) expr() (ast.Node, error)    { return p.orExpr() }

func (p *Parser) orExpr() (ast.Node, error) {
	left, err := p.andExpr()
	if err != nil {
		return nil, err
	}
	for p.current().Type == lexer.TOKEN_OR {
		p.pos++
		right, err := p.andExpr()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryOp{Left: left, Op: "or", Right: right}
	}
	return left, nil
}

func (p *Parser) andExpr() (ast.Node, error) {
	left, err := p.notExpr()
	if err != nil {
		return nil, err
	}
	for p.current().Type == lexer.TOKEN_AND {
		p.pos++
		right, err := p.notExpr()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryOp{Left: left, Op: "and", Right: right}
	}
	return left, nil
}

func (p *Parser) notExpr() (ast.Node, error) {
	if p.current().Type == lexer.TOKEN_NOT {
		tok := p.current()
		p.pos++
		operand, err := p.notExpr()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryOp{Op: "not", Operand: operand, Line: tok.Line}, nil
	}
	return p.comparison()
}

func (p *Parser) comparison() (ast.Node, error) {
	left, err := p.arithExpr()
	if err != nil {
		return nil, err
	}
	for p.current().Type == lexer.TOKEN_EQEQ || p.current().Type == lexer.TOKEN_NEQ ||
		p.current().Type == lexer.TOKEN_LT || p.current().Type == lexer.TOKEN_GT ||
		p.current().Type == lexer.TOKEN_LTE || p.current().Type == lexer.TOKEN_GTE {
		op := p.current().Value
		p.pos++
		right, err := p.arithExpr()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryOp{Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) arithExpr() (ast.Node, error) {
	left, err := p.term()
	if err != nil {
		return nil, err
	}
	for p.current().Type == lexer.TOKEN_PLUS || p.current().Type == lexer.TOKEN_MINUS {
		op := p.current().Value
		p.pos++
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryOp{Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) term() (ast.Node, error) {
	left, err := p.unary()
	if err != nil {
		return nil, err
	}
	for p.current().Type == lexer.TOKEN_STAR || p.current().Type == lexer.TOKEN_SLASH || p.current().Type == lexer.TOKEN_MOD {
		op := p.current().Value
		p.pos++
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		left = &ast.BinaryOp{Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) unary() (ast.Node, error) {
	if p.current().Type == lexer.TOKEN_MINUS {
		tok := p.current()
		p.pos++
		operand, err := p.unary()
		if err != nil {
			return nil, err
		}
		return &ast.UnaryOp{Op: "-", Operand: operand, Line: tok.Line}, nil
	}
	return p.postfix()
}

func (p *Parser) postfix() (ast.Node, error) {
	node, err := p.atom()
	if err != nil {
		return nil, err
	}
	for {
		if p.current().Type == lexer.TOKEN_LPAREN {
			ident, ok := node.(*ast.Identifier)
			if !ok {
				return nil, fmt.Errorf("[SyntaxError] (line %d): Can only call functions by name", p.current().Line)
			}
			p.eat(lexer.TOKEN_LPAREN)
			args, err := p.argList()
			if err != nil {
				return nil, err
			}
			p.eat(lexer.TOKEN_RPAREN)
			node = &ast.FuncCall{Name: ident.Name, Args: args, Line: ident.Line}
		} else if p.current().Type == lexer.TOKEN_LBRACKET {
			p.eat(lexer.TOKEN_LBRACKET)
			index, err := p.expr()
			if err != nil {
				return nil, err
			}
			p.eat(lexer.TOKEN_RBRACKET)
			node = &ast.IndexAccess{Object: node, Index: index}
		} else if p.current().Type == lexer.TOKEN_DOT {
			p.eat(lexer.TOKEN_DOT)
			methodTok, err := p.eat(lexer.TOKEN_IDENT)
			if err != nil {
				return nil, err
			}
			if p.current().Type == lexer.TOKEN_LPAREN {
				p.eat(lexer.TOKEN_LPAREN)
				args, err := p.argList()
				if err != nil {
					return nil, err
				}
				p.eat(lexer.TOKEN_RPAREN)
				node = &ast.MethodCall{Object: node, Method: methodTok.Value, Args: args, Line: methodTok.Line}
			} else {
				node = &ast.MethodCall{Object: node, Method: methodTok.Value, Args: []ast.Node{}, Line: methodTok.Line}
			}
		} else {
			break
		}
	}
	return node, nil
}

func (p *Parser) argList() ([]ast.Node, error) {
	args := []ast.Node{}
	if p.current().Type == lexer.TOKEN_RPAREN {
		return args, nil
	}
	arg, err := p.expr()
	if err != nil {
		return nil, err
	}
	args = append(args, arg)
	for p.current().Type == lexer.TOKEN_COMMA {
		p.eat(lexer.TOKEN_COMMA)
		arg, err := p.expr()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}
	return args, nil
}

func (p *Parser) atom() (ast.Node, error) {
	tok := p.current()
	switch tok.Type {
	case lexer.TOKEN_INT:
		p.pos++
		return &ast.NumberLiteral{Value: tok.Value, IsFloat: false, Line: tok.Line}, nil
	case lexer.TOKEN_FLOAT:
		p.pos++
		return &ast.NumberLiteral{Value: tok.Value, IsFloat: true, Line: tok.Line}, nil
	case lexer.TOKEN_STRING:
		p.pos++
		return &ast.StringLiteral{Value: tok.Value, Line: tok.Line}, nil
	case lexer.TOKEN_KEYWORD:
		switch tok.Value {
		case "true":
			p.pos++
			return &ast.BoolLiteral{Value: true, Line: tok.Line}, nil
		case "false":
			p.pos++
			return &ast.BoolLiteral{Value: false, Line: tok.Line}, nil
		case "null":
			p.pos++
			return &ast.NullLiteral{Line: tok.Line}, nil
		}
	case lexer.TOKEN_IDENT:
		p.pos++
		return &ast.Identifier{Name: tok.Value, Line: tok.Line}, nil
	case lexer.TOKEN_LPAREN:
		p.eat(lexer.TOKEN_LPAREN)
		e, err := p.expr()
		if err != nil {
			return nil, err
		}
		p.eat(lexer.TOKEN_RPAREN)
		return e, nil
	case lexer.TOKEN_LBRACKET:
		return p.listLiteral()
	}
	return nil, fmt.Errorf("[SyntaxError] (line %d): Unexpected token '%s'", tok.Line, tok.Value)
}

func (p *Parser) listLiteral() (ast.Node, error) {
	tok, _ := p.eat(lexer.TOKEN_LBRACKET)
	elements := []ast.Node{}
	p.skipNewlines()
	if p.current().Type != lexer.TOKEN_RBRACKET {
		el, err := p.expr()
		if err != nil {
			return nil, err
		}
		elements = append(elements, el)
		for p.current().Type == lexer.TOKEN_COMMA {
			p.eat(lexer.TOKEN_COMMA)
			p.skipNewlines()
			el, err := p.expr()
			if err != nil {
				return nil, err
			}
			elements = append(elements, el)
		}
	}
	p.skipNewlines()
	p.eat(lexer.TOKEN_RBRACKET)
	return &ast.ListLiteral{Elements: elements, Line: tok.Line}, nil
}
