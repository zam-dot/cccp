// [file name]: parser.go
// [file content begin]
package parser

import (
	"cccp/ast"
	"cccp/lexer"
	"fmt"
	"strconv"
)

// Precedence levels for operator parsing (Pratt parsing).
// Higher values indicate higher precedence.
const (
	_           int = iota
	LOWEST          // Lowest precedence
	EQUALS          // ==, !=
	LESSGREATER     // >, <        // Reserved for future comparison operators
	SUM             // +, -
	PRODUCT         // *, /
	PREFIX          // -X or !X    // Unary operators
	CALL            // myFunction(X) // Function calls
)

// precedences maps token types to their precedence levels for infix operators.
var precedences = map[ast.TokenType]int{
	ast.EQ:       EQUALS,  // == operator
	ast.NOT_EQ:   EQUALS,  // != operator
	ast.PLUS:     SUM,     // + operator
	ast.MINUS:    SUM,     // - operator
	ast.ASTERISK: PRODUCT, // * operator
	ast.SLASH:    PRODUCT, // / operator
}

// Parser processes tokens and constructs an Abstract Syntax Tree.
// It uses Pratt parsing for handling operator precedence.
type Parser struct {
	l      *lexer.Lexer // The lexer providing tokens
	errors []string     // Collection of parsing errors

	curToken  *ast.Token // Current token being examined
	peekToken *ast.Token // Next token lookahead

	// Function tables for prefix and infix parsing
	prefixParseFns map[ast.TokenType]func() ast.Expression
	infixParseFns  map[ast.TokenType]func(ast.Expression) ast.Expression
}

// New creates a new Parser with the given lexer.
// Initializes the parser and sets up parsing function registries.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// Initialize prefix parsing function registry
	p.prefixParseFns = make(map[ast.TokenType]func() ast.Expression)
	p.registerPrefix(ast.IDENT, p.parseIdentifier)
	p.registerPrefix(ast.INT, p.parseIntegerLiteral)
	p.registerPrefix(ast.STRING, p.parseStringLiteral)
	p.registerPrefix(ast.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(ast.FUNC, p.parseFunctionLiteral) // Function expressions

	// Initialize infix parsing function registry
	p.infixParseFns = make(map[ast.TokenType]func(ast.Expression) ast.Expression)
	p.registerInfix(ast.PLUS, p.parseInfixExpression)
	p.registerInfix(ast.MINUS, p.parseInfixExpression)
	p.registerInfix(ast.ASTERISK, p.parseInfixExpression)
	p.registerInfix(ast.SLASH, p.parseInfixExpression)
	p.registerInfix(ast.EQ, p.parseInfixExpression)
	p.registerInfix(ast.NOT_EQ, p.parseInfixExpression)

	// Initialize tokens by reading twice to set curToken and peekToken
	p.nextToken() // sets curToken
	p.nextToken() // sets peekToken

	return p
}

// parseFunctionLiteral parses a function literal expression.
// Example: func(x, y) { return x + y; }
func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(ast.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(ast.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

// parseFunctionParameters parses the parameter list of a function.
// Returns a list of identifier nodes representing the parameters.
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	// Handle empty parameter list: func()
	if p.peekTokenIs(ast.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	// Parse first parameter
	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	// Parse additional parameters separated by commas
	for p.peekTokenIs(ast.COMMA) {
		p.nextToken() // skip comma
		p.nextToken() // move to next parameter
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(ast.RPAREN) {
		return nil
	}

	return identifiers
}

// parseReturnStatement parses a return statement.
// Example: return 5; or return x + y;
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	// Return can have an optional expression
	if !p.curTokenIs(ast.SEMICOLON) {
		stmt.ReturnValue = p.parseExpression(LOWEST)
	}

	// Optional semicolon
	if p.peekTokenIs(ast.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseStringLiteral parses a string literal expression.
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}
}

// ParseProgram is the main entry point for parsing.
// It parses the entire program and returns the root AST node.
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	// Parse all statements until EOF
	for p.curToken.Type != ast.EOF {
		fmt.Printf("🔍 DEBUG ParseProgram: parsing statement at %s '%s'\n",
			p.curToken.Type, p.curToken.Literal)

		stmt := p.parseStatement()
		if stmt != nil {
			fmt.Printf("✅ DEBUG ParseProgram: added statement: %T\n", stmt)
			program.Statements = append(program.Statements, stmt)
		} else {
			fmt.Printf("❌ DEBUG ParseProgram: parseStatement returned nil at %s '%s'\n",
				p.curToken.Type, p.curToken.Literal)
		}

		// Only advance if we're not at EOF
		if p.curToken.Type != ast.EOF {
			p.nextToken()
			fmt.Printf("🔍 DEBUG ParseProgram: advanced to %s '%s'\n",
				p.curToken.Type, p.curToken.Literal)
		}
	}

	fmt.Println("✅ DEBUG ParseProgram: finished parsing program")
	return program
}

// Errors returns the collection of parsing errors encountered.
func (p *Parser) Errors() []string {
	return p.errors
}

// parseStatement is the main dispatch function for parsing statements.
// It determines the type of statement based on the current token and delegates to appropriate parsers.
func (p *Parser) parseStatement() ast.Statement {
	fmt.Printf(
		"🚀 DEBUG parseStatement: current token = %s '%s'\n",
		p.curToken.Type,
		p.curToken.Literal,
	)

	switch p.curToken.Type {
	case ast.VAR:
		fmt.Println("🔍 DEBUG: Parsing VAR statement")
		return p.parseLetStatement()
	case ast.PRINT:
		fmt.Println("🔍 DEBUG: Parsing PRINT statement")
		return p.parsePrintStatement()
	case ast.IF:
		fmt.Println("🔍 DEBUG: Parsing IF statement")
		return p.parseIfStatement()
	case ast.EXTERN:
		fmt.Println("🔍 DEBUG: Parsing EXTERN statement")
		return p.parseExternStatement()
	case ast.RETURN:
		fmt.Println("🔍 DEBUG: Parsing RETURN statement")
		return p.parseReturnStatement()
	case ast.FUNC:
		fmt.Println("🔍 DEBUG: Parsing FUNC statement")
		return p.parseFunctionStatement()
	case ast.LBRACE:
		fmt.Println("🔍 DEBUG: Parsing block statement")
		return p.parseBlockStatement()
	case ast.IDENT:
		fmt.Println("🔍 DEBUG: Parsing IDENT statement")
		if p.peekTokenIs(ast.ASSIGN) {
			return p.parseAssignmentStatement()
		}
		// Handle function calls as expressions at statement level
		fmt.Println("🔍 DEBUG: Trying to parse as expression statement")
		expr := p.parseExpression(LOWEST)
		fmt.Printf("✅ DEBUG: parseExpression returned: %T\n", expr)
		if expr != nil {
			fmt.Println("✅ DEBUG: Creating ExpressionStatement")
			return &ast.ExpressionStatement{
				Token:      p.curToken,
				Expression: expr,
			}
		}
		fallthrough
	default:
		fmt.Printf("❌ DEBUG: Unknown token type in parseStatement: %s\n", p.curToken.Type)
		return nil
	}
}

// parseFunctionStatement parses a named function definition.
// Example: func add(a, b) { return a + b; }
func (p *Parser) parseFunctionStatement() *ast.FunctionStatement {
	fmt.Println("🔍 DEBUG: Starting parseFunctionStatement")
	stmt := &ast.FunctionStatement{Token: p.curToken}

	// Function name
	if !p.expectPeek(ast.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	fmt.Printf("✅ DEBUG: Function name: %s\n", stmt.Name.Value)

	// Parameters
	if !p.expectPeek(ast.LPAREN) {
		return nil
	}
	stmt.Parameters = p.parseFunctionParameters()
	fmt.Printf("✅ DEBUG: Function has %d parameters\n", len(stmt.Parameters))

	// Function body - expect '{'
	if !p.expectPeek(ast.LBRACE) {
		return nil
	}

	// Parse the entire function block
	stmt.Body = p.parseBlockStatement()

	// DEBUG: Print all statements in function body
	fmt.Printf(
		"✅ DEBUG: Function %s body has %d statements:\n",
		stmt.Name.Value,
		len(stmt.Body.Statements),
	)
	for i, bodyStmt := range stmt.Body.Statements {
		fmt.Printf("  %d: %T\n", i, bodyStmt)
	}

	return stmt
}

// parseExternStatement parses an external function declaration.
// Example: extern printf;
func (p *Parser) parseExternStatement() *ast.ExternStatement {
	fmt.Println("🔍 DEBUG: In parseExternStatement")
	stmt := &ast.ExternStatement{Token: p.curToken}

	fmt.Printf("🔍 DEBUG: Current token: %s '%s', Peek token: %s '%s'\n",
		p.curToken.Type, p.curToken.Literal,
		p.peekToken.Type, p.peekToken.Literal)

	if !p.expectPeek(ast.IDENT) {
		fmt.Println("❌ DEBUG: Failed to expect IDENT after EXTERN")
		return nil
	}

	// Function name
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	fmt.Printf("✅ DEBUG: Extern function name: %s\n", stmt.Name.Value)

	// For simple syntax: extern printf;
	if p.peekTokenIs(ast.SEMICOLON) {
		fmt.Println("✅ DEBUG: Found semicolon - simple extern declaration")
		p.nextToken()
		return stmt
	}

	fmt.Println("❌ DEBUG: Extern with parameters not implemented yet")
	return nil
}

// parseIfStatement parses an if statement with optional else clause.
// Example: if (x > 0) { print("positive"); }
func (p *Parser) parseIfStatement() *ast.IfStatement {
	fmt.Println("🔍 DEBUG: Starting parseIfStatement")
	stmt := &ast.IfStatement{Token: p.curToken}

	// Move to the condition
	p.nextToken()
	fmt.Printf("🔍 DEBUG: Before parsing condition: %s '%s'\n",
		p.curToken.Type, p.curToken.Literal)

	// Parse the condition
	stmt.Condition = p.parseExpression(LOWEST)
	fmt.Printf("🔍 DEBUG: Parsed condition: %T\n", stmt.Condition)

	// Expect '{' after condition
	if !p.expectPeek(ast.LBRACE) {
		fmt.Printf("❌ DEBUG: Expected { after if condition, got %s '%s'\n",
			p.peekToken.Type, p.peekToken.Literal)
		return nil
	}

	// Parse the consequence block
	stmt.Consequence = p.parseBlockStatement()
	fmt.Printf("🔍 DEBUG: Parsed consequence with %d statements\n",
		len(stmt.Consequence.Statements))

	stmt.Alternative = nil

	// CRITICAL FIX: Don't consume any tokens after the block!
	// The block parsing should leave us positioned at the next statement
	fmt.Printf("🔍 DEBUG: parseIfStatement finished at %s '%s'\n",
		p.curToken.Type, p.curToken.Literal)

	fmt.Println("✅ DEBUG: parseIfStatement completed successfully")
	return stmt
}

// parseBlockStatement parses a block of statements enclosed in braces.
// Example: { x = 1; y = 2; }
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	fmt.Println("🔍 DEBUG: Starting parseBlockStatement")
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	braceCount := 1 // Start with 1 for the opening brace

	// Parse until we find the matching closing brace
	for braceCount > 0 && !p.curTokenIs(ast.EOF) {
		fmt.Printf("🔍 DEBUG: parseBlockStatement parsing: %s '%s' (braceCount=%d)\n",
			p.curToken.Type, p.curToken.Literal, braceCount)

		if p.curTokenIs(ast.LBRACE) {
			braceCount++
		} else if p.curTokenIs(ast.RBRACE) {
			braceCount--
			if braceCount == 0 {
				// This is the matching closing brace for our block
				break
			}
		}

		// Only parse statements if we're not at a closing brace
		if !p.curTokenIs(ast.RBRACE) || braceCount > 0 {
			stmt := p.parseStatement()
			if stmt != nil {
				fmt.Printf("✅ DEBUG: Added statement to block: %T\n", stmt)
				block.Statements = append(block.Statements, stmt)
			}
		}

		// Only advance if we're not at the end
		if braceCount > 0 && !p.curTokenIs(ast.EOF) {
			p.nextToken()
		}
	}

	fmt.Printf("🔍 DEBUG: Finished parseBlockStatement with %d statements, braceCount=%d\n",
		len(block.Statements), braceCount)
	return block
}

// parseAssignmentStatement parses a variable assignment.
// Example: x = 10;
func (p *Parser) parseAssignmentStatement() *ast.AssignmentStatement {
	stmt := &ast.AssignmentStatement{}

	// The current token is the identifier (variable name)
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// The next token should be '='
	if !p.expectPeek(ast.ASSIGN) {
		return nil
	}

	// Move to the value after '=' and parse it as an expression
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	// Optional semicolon
	if p.peekTokenIs(ast.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseLetStatement parses a variable declaration.
// Example: var x = 5; or var x;
func (p *Parser) parseLetStatement() *ast.LetStatement {
	fmt.Println("🔍 DEBUG: Starting parseLetStatement")
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(ast.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	fmt.Printf("🔍 DEBUG: Variable name: %s\n", stmt.Name.Value)

	// Check if there's an assignment or just a declaration
	if p.peekTokenIs(ast.SEMICOLON) {
		// No assignment, just declaration: "var x;"
		fmt.Println("🔍 DEBUG: No assignment, just declaration")
		p.nextToken()
		return stmt
	} else if p.peekTokenIs(ast.ASSIGN) {
		// There's an assignment: "var x = 5;"
		fmt.Println("🔍 DEBUG: Found assignment")
		p.nextToken() // skip '='
		p.nextToken() // move to the value

		fmt.Printf("🔍 DEBUG: Current token before parseExpression: %s '%s'\n",
			p.curToken.Type, p.curToken.Literal)
		fmt.Printf("🔍 DEBUG: Peek token before parseExpression: %s '%s'\n",
			p.peekToken.Type, p.peekToken.Literal)

		// Use parseExpression instead of assuming it's a simple value
		stmt.Value = p.parseExpression(LOWEST)
		fmt.Printf("🔍 DEBUG: parseExpression returned: %T\n", stmt.Value)
	}

	// Optional semicolon
	if p.peekTokenIs(ast.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parsePrintStatement parses a print statement.
// Example: print("Hello world");
func (p *Parser) parsePrintStatement() *ast.PrintStatement {
	fmt.Println("🔍 DEBUG: Starting parsePrintStatement")
	stmt := &ast.PrintStatement{Token: p.curToken}

	// Expect parentheses after print
	fmt.Printf("🔍 DEBUG: Current: %s '%s', Peek: %s '%s'\n",
		p.curToken.Type, p.curToken.Literal,
		p.peekToken.Type, p.peekToken.Literal)

	if !p.expectPeek(ast.LPAREN) {
		fmt.Println("❌ DEBUG: Expected ( after print")
		return nil
	}

	p.nextToken() // move past (
	fmt.Printf("🔍 DEBUG: After moving past (: %s '%s'\n",
		p.curToken.Type, p.curToken.Literal)

	stmt.Value = p.parseExpression(LOWEST)
	fmt.Printf("🔍 DEBUG: Print value parsed: %T\n", stmt.Value)

	// Expect closing )
	if !p.expectPeek(ast.RPAREN) {
		fmt.Printf("❌ DEBUG: Expected ) after print value, got %s '%s'\n",
			p.peekToken.Type, p.peekToken.Literal)
		return nil
	}
	p.nextToken() // move past )

	// Check if there's a semicolon (optional)
	if p.peekTokenIs(ast.SEMICOLON) {
		p.nextToken()
	}

	fmt.Printf("🔍 DEBUG: parsePrintStatement finished at: %s '%s'\n",
		p.curToken.Type, p.curToken.Literal)
	fmt.Println("✅ DEBUG: parsePrintStatement completed successfully")
	return stmt
}

// parseExpression is the core Pratt parser for expressions.
// It handles operator precedence by recursively parsing expressions with higher precedence first.
func (p *Parser) parseExpression(precedence int) ast.Expression {
	fmt.Printf("🔍 DEBUG parseExpression: current=%s '%s', precedence=%d\n",
		p.curToken.Type, p.curToken.Literal, precedence)

	// Parse prefix expression (left-hand side)
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	fmt.Printf("🔍 DEBUG: After prefix parse, leftExp=%T, peek=%s '%s'\n",
		leftExp, p.peekToken.Type, p.peekToken.Literal)

	// Handle function calls after identifiers
	for p.peekTokenIs(ast.LPAREN) {
		fmt.Println("🔍 DEBUG: Found LPAREN after expression - parsing function call")
		p.nextToken() // consume the '('
		leftExp = p.parseFunctionCall(leftExp)
		fmt.Printf("🔍 DEBUG: After function call parse, leftExp=%T\n", leftExp)
	}

	// Handle infix operators with higher precedence
	for !p.peekTokenIs(ast.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

// parseFunctionCall parses a function call expression.
// Example: add(1, 2)
func (p *Parser) parseFunctionCall(function ast.Expression) ast.Expression {
	fmt.Printf("🔍 DEBUG parseFunctionCall: function=%T\n", function)
	if function == nil {
		return nil
	}

	call := &ast.FunctionCall{Token: p.curToken, Function: function}
	call.Arguments = p.parseExpressionList(ast.RPAREN)
	fmt.Printf("🔍 DEBUG: Function call has %d arguments\n", len(call.Arguments))
	return call
}

// parseExpressionList parses a comma-separated list of expressions.
// Used for function arguments and other list contexts.
func (p *Parser) parseExpressionList(end ast.TokenType) []ast.Expression {
	fmt.Printf("🔍 DEBUG parseExpressionList: current token = %s '%s', end = %s\n",
		p.curToken.Type, p.curToken.Literal, end)

	list := []ast.Expression{}

	// Handle empty list
	if p.peekTokenIs(end) {
		fmt.Println("✅ DEBUG: Empty expression list")
		p.nextToken()
		return list
	}

	p.nextToken()
	fmt.Printf("🔍 DEBUG: Parsing first expression: %s '%s'\n",
		p.curToken.Type, p.curToken.Literal)

	// Parse first expression
	expr := p.parseExpression(LOWEST)
	if expr != nil {
		fmt.Printf("✅ DEBUG: Added expression: %T\n", expr)
		list = append(list, expr)
	}

	// Parse additional expressions separated by commas
	for p.peekTokenIs(ast.COMMA) {
		fmt.Println("🔍 DEBUG: Found comma, parsing next expression")
		p.nextToken() // skip comma
		p.nextToken() // move to next expression
		expr := p.parseExpression(LOWEST)
		if expr != nil {
			fmt.Printf("✅ DEBUG: Added expression: %T\n", expr)
			list = append(list, expr)
		}
	}

	if !p.expectPeek(end) {
		fmt.Printf("❌ DEBUG: Expected %s but got %s\n", end, p.peekToken.Type)
		return nil
	}

	fmt.Printf("✅ DEBUG: parseExpressionList returning %d expressions\n", len(list))
	return list
}

// parseIdentifier parses an identifier expression.
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseIntegerLiteral parses an integer literal expression.
func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

// parseInfixExpression parses an infix (binary) operator expression.
// Examples: x + y, a * b, foo == bar
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

// parseGroupedExpression parses an expression enclosed in parentheses.
// Example: (x + y) * z
func (p *Parser) parseGroupedExpression() ast.Expression {
	fmt.Println("🔍 DEBUG: Starting parseGroupedExpression")
	p.nextToken()

	exp := p.parseExpression(LOWEST)
	fmt.Printf("🔍 DEBUG: parseGroupedExpression parsed: %T\n", exp)

	if !p.expectPeek(ast.RPAREN) {
		fmt.Println("❌ DEBUG: Expected ) in grouped expression")
		return nil
	}

	return exp
}

// Helper methods for parser state management...

// nextToken advances the parser to the next token.
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// curTokenIs checks if the current token is of the given type.
func (p *Parser) curTokenIs(t ast.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs checks if the next token is of the given type.
func (p *Parser) peekTokenIs(t ast.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek checks if the next token is of the expected type and advances if it is.
// Returns true if the token matches, false otherwise (and records an error).
func (p *Parser) expectPeek(t ast.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// peekError records an error for an unexpected peek token.
func (p *Parser) peekError(t ast.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

// registerPrefix registers a prefix parsing function for a token type.
func (p *Parser) registerPrefix(tokenType ast.TokenType, fn func() ast.Expression) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix registers an infix parsing function for a token type.
func (p *Parser) registerInfix(tokenType ast.TokenType, fn func(ast.Expression) ast.Expression) {
	p.infixParseFns[tokenType] = fn
}

// curPrecedence returns the precedence of the current token.
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

// peekPrecedence returns the precedence of the next token.
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

// noPrefixParseFnError records an error for a token with no prefix parsing function.
func (p *Parser) noPrefixParseFnError(t ast.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// [file content end]
