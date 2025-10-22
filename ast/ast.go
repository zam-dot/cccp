// [file name]: ast.go
// [file content begin]
package ast

// Node is the interface that all AST nodes must implement.
// It provides the basic contract for all nodes in the Abstract Syntax Tree.
type Node interface {
	// TokenLiteral returns the literal value of the token associated with this node
	TokenLiteral() string
}

// Statement nodes represent actions that do not produce a value.
// Statements are executed for their side effects.
type Statement interface {
	Node
	statementNode() // Marker method to distinguish statements from expressions
}

// Expression nodes represent pieces of code that produce a value.
// Expressions can be evaluated to yield a result.
type Expression interface {
	Node
	expressionNode() // Marker method to distinguish expressions from statements
}

// Program is the root node of every AST.
// It contains all the top-level statements in the program.
type Program struct {
	Statements []Statement
}

// TokenLiteral returns the token literal of the first statement in the program.
// If the program is empty, returns an empty string.
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// LetStatement represents a variable assignment using the 'var' keyword.
// Example: var x = 5;
type LetStatement struct {
	Token *Token      // The 'var' token
	Name  *Identifier // The variable name being declared
	Value Expression  // The expression being assigned to the variable
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

// Identifier represents a variable name in the source code.
type Identifier struct {
	Token *Token // The IDENT token
	Value string // The actual name of the identifier
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// PrintStatement represents a print call to output values.
// Example: print("Hello world");
type PrintStatement struct {
	Token *Token     // The 'print' token
	Value Expression // The expression to be printed
}

func (ps *PrintStatement) statementNode()       {}
func (ps *PrintStatement) TokenLiteral() string { return ps.Token.Literal }

// IntegerLiteral represents an integer value in the source code.
type IntegerLiteral struct {
	Token *Token
	Value int64 // The parsed integer value
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }

// InfixExpression represents a binary operation with an operator between two expressions.
// Example: x + y, a * b
type InfixExpression struct {
	Token    *Token     // The operator token, e.g., '+'
	Left     Expression // The left-hand side expression
	Operator string     // The operator as a string
	Right    Expression // The right-hand side expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }

// Token represents a single lexical token parsed from the source code.
type Token struct {
	Type    TokenType // The type of token (e.g., IDENT, INT, PLUS)
	Literal string    // The actual text of the token from source
}

// StringLiteral represents a string value in the source code.
type StringLiteral struct {
	Token *Token
	Value string // The string value without quotes
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }

// TokenType represents the type of a lexical token.
type TokenType string

// Define our core token types.
const (
	// Comparison operators
	EQ     TokenType = "=="
	NOT_EQ TokenType = "!="

	// Special tokens
	ILLEGAL TokenType = "ILLEGAL" // Invalid token
	EOF     TokenType = "EOF"     // End of file

	// Identifiers and basic literals
	IDENT  TokenType = "IDENT"  // Variable/function names
	INT    TokenType = "INT"    // Integer literals
	STRING TokenType = "STRING" // String literals

	// Operators
	ASSIGN TokenType = "=" // Assignment operator
	PLUS   TokenType = "+" // Addition operator

	// Delimiters
	COMMA     TokenType = "," // Comma separator
	SEMICOLON TokenType = ";" // Statement terminator
	LPAREN    TokenType = "(" // Left parenthesis
	RPAREN    TokenType = ")" // Right parenthesis

	// Keywords
	PRINT TokenType = "PRINT" // Print keyword
	VAR   TokenType = "VAR"   // Variable declaration keyword

	// Math operators
	MINUS    TokenType = "-" // Subtraction operator
	ASTERISK TokenType = "*" // Multiplication operator
	SLASH    TokenType = "/" // Division operator

	// Control flow
	IF     TokenType = "IF" // If statement keyword
	LBRACE TokenType = "{"  // Left brace for blocks
	RBRACE TokenType = "}"  // Right brace for blocks

	// External functions
	EXTERN TokenType = "EXTERN" // External function declaration

	// Additional syntax
	COLON    TokenType = ":"   // Colon (for type annotations)
	DOT      TokenType = "."   // Dot operator
	ELLIPSIS TokenType = "..." // Ellipsis for varargs

	// Function-related
	FUNC   TokenType = "FUNC"   // Function definition
	RETURN TokenType = "RETURN" // Return statement
)

// AssignmentStatement represents a variable reassignment.
// Example: x = 10;
type AssignmentStatement struct {
	Token *Token      // The '=' token
	Name  *Identifier // The variable being assigned to
	Value Expression  // The new value
}

func (as *AssignmentStatement) statementNode()       {}
func (as *AssignmentStatement) TokenLiteral() string { return as.Token.Literal }

// IfStatement represents a conditional branch in the program.
// Example: if (x > 0) { print("positive"); }
type IfStatement struct {
	Token       *Token          // The 'if' token
	Condition   Expression      // The condition to evaluate
	Consequence *BlockStatement // The block to execute if condition is true
	Alternative *BlockStatement // The else clause (optional)
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }

// BlockStatement represents a sequence of statements enclosed in braces.
// Example: { x = 1; y = 2; }
type BlockStatement struct {
	Token      *Token      // The { token
	Statements []Statement // The statements within the block
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }

// ExternStatement declares an external function (typically from C).
// Example: extern printf;
type ExternStatement struct {
	Token      *Token        // The 'extern' token
	Name       *Identifier   // The external function name
	Params     []*Identifier // Parameter names
	ReturnType *Identifier   // Return type identifier
}

func (es *ExternStatement) statementNode()       {}
func (es *ExternStatement) TokenLiteral() string { return es.Token.Literal }

// FunctionCall represents a function invocation.
// Example: add(1, 2)
type FunctionCall struct {
	Token     *Token       // The function name token
	Function  Expression   // Identifier or FunctionLiteral
	Arguments []Expression // The arguments passed to the function
}

func (fc *FunctionCall) expressionNode()      {}
func (fc *FunctionCall) TokenLiteral() string { return fc.Token.Literal }

// ExpressionStatement wraps an expression as a statement.
// Used when an expression appears at statement level.
type ExpressionStatement struct {
	Token      *Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

// FunctionLiteral represents an anonymous function definition.
// Example: func(x, y) { return x + y; }
type FunctionLiteral struct {
	Token      *Token          // The 'func' token
	Parameters []*Identifier   // Parameter list
	Body       *BlockStatement // Function body
}

func (fl *FunctionLiteral) expressionNode()      {}
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }

// ReturnStatement represents a return statement in a function.
// Example: return x + y;
type ReturnStatement struct {
	Token       *Token     // The 'return' token
	ReturnValue Expression // The value to return (optional)
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

// FunctionStatement represents a named function definition.
// Example: func add(a, b) { return a + b; }
type FunctionStatement struct {
	Token      *Token          // The 'func' token
	Name       *Identifier     // Function name
	Parameters []*Identifier   // Parameter list
	Body       *BlockStatement // Function body
}

func (fs *FunctionStatement) statementNode()       {}
func (fs *FunctionStatement) TokenLiteral() string { return fs.Token.Literal }

// [file content end]
