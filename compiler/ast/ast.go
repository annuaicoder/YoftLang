package ast

type Node interface {
	nodeType() string
}

type Program struct {
	Statements []Node
}

func (p *Program) nodeType() string { return "Program" }

type NumberLiteral struct {
	Value string
	IsFloat bool
	Line  int
}

func (n *NumberLiteral) nodeType() string { return "NumberLiteral" }

type StringLiteral struct {
	Value string
	Line  int
}

func (n *StringLiteral) nodeType() string { return "StringLiteral" }

type BoolLiteral struct {
	Value bool
	Line  int
}

func (n *BoolLiteral) nodeType() string { return "BoolLiteral" }

type NullLiteral struct {
	Line int
}

func (n *NullLiteral) nodeType() string { return "NullLiteral" }

type ListLiteral struct {
	Elements []Node
	Line     int
}

func (n *ListLiteral) nodeType() string { return "ListLiteral" }

type Identifier struct {
	Name string
	Line int
}

func (n *Identifier) nodeType() string { return "Identifier" }

type VarDecl struct {
	Name  string
	Value Node
	Line  int
}

func (n *VarDecl) nodeType() string { return "VarDecl" }

type VarReassign struct {
	Name  string
	Value Node
	Line  int
}

func (n *VarReassign) nodeType() string { return "VarReassign" }

type BinaryOp struct {
	Left  Node
	Op    string
	Right Node
	Line  int
}

func (n *BinaryOp) nodeType() string { return "BinaryOp" }

type UnaryOp struct {
	Op      string
	Operand Node
	Line    int
}

func (n *UnaryOp) nodeType() string { return "UnaryOp" }

type ShowStmt struct {
	Value Node
	Line  int
}

func (n *ShowStmt) nodeType() string { return "ShowStmt" }

type IfStmt struct {
	Condition Node
	Body      []Node
	ElseBody  []Node
	Line      int
}

func (n *IfStmt) nodeType() string { return "IfStmt" }

type WhileStmt struct {
	Condition Node
	Body      []Node
	Line      int
}

func (n *WhileStmt) nodeType() string { return "WhileStmt" }

type ForStmt struct {
	VarName  string
	Iterable Node
	Body     []Node
	Line     int
}

func (n *ForStmt) nodeType() string { return "ForStmt" }

type FuncDecl struct {
	Name   string
	Params []string
	Body   []Node
	Line   int
}

func (n *FuncDecl) nodeType() string { return "FuncDecl" }

type FuncCall struct {
	Name string
	Args []Node
	Line int
}

func (n *FuncCall) nodeType() string { return "FuncCall" }

type ReturnStmt struct {
	Value Node
	Line  int
}

func (n *ReturnStmt) nodeType() string { return "ReturnStmt" }

type IndexAccess struct {
	Object Node
	Index  Node
	Line   int
}

func (n *IndexAccess) nodeType() string { return "IndexAccess" }

type MethodCall struct {
	Object Node
	Method string
	Args   []Node
	Line   int
}

func (n *MethodCall) nodeType() string { return "MethodCall" }

type ImportStmt struct {
	Path string
	Line int
}

func (n *ImportStmt) nodeType() string { return "ImportStmt" }
