// [file name]: lexer.go
// [file content begin]
package lexer

import (
	"cccp/ast" // Update with your actual module name
)

// Lexer is responsible for converting source code into tokens.
// It reads characters sequentially and groups them into meaningful tokens.
type Lexer struct {
	input        string // The source code being lexed
	position     int    // Current position in input (points to current char)
	readPosition int    // Next reading position in input (after current char)
	ch           byte   // Current character under examination
}

// New creates a new Lexer from the input string and initializes it.
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar() // Initialize by reading the first character
	return l
}

// NextToken parses and returns the next token from the input.
// It skips whitespace and comments, then identifies the next token based on the current character.
func (l *Lexer) NextToken() *ast.Token {
	var tok ast.Token

	l.skipWhitespace()

	// Handle comments before processing tokens
	if l.ch == '/' && l.peekChar() == '/' {
		l.skipLineComment()
		return l.NextToken() // Return next token after comment
	}
	if l.ch == '/' && l.peekChar() == '*' {
		l.skipBlockComment()
		return l.NextToken() // Return next token after comment
	}

	switch l.ch {
	case '"': // String literal
		tok.Literal = l.readString()
		tok.Type = ast.STRING
		return &tok

	case '=':
		if l.peekChar() == '=' {
			// It's == (equality operator)
			l.readChar()
			tok = ast.Token{Type: ast.EQ, Literal: "=="}
		} else {
			tok = newToken(ast.ASSIGN, l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			// It's != (inequality operator)
			l.readChar()
			tok = ast.Token{Type: ast.NOT_EQ, Literal: "!="}
		} else {
			// Handle other ! cases later (like logical NOT)
			tok = newToken(ast.ILLEGAL, l.ch)
		}
	case ':': // Colon (for type annotations)
		tok = newToken(ast.COLON, l.ch)
	case '.': // Dot operator or ellipsis
		// Check for ellipsis '...'
		if l.peekChar() == '.' {
			l.readChar()
			if l.peekChar() == '.' {
				l.readChar()
				tok = ast.Token{Type: ast.ELLIPSIS, Literal: "..."}
			} else {
				// Handle error case for incomplete ellipsis
				tok = newToken(ast.ILLEGAL, l.ch)
			}
		} else {
			tok = newToken(ast.DOT, l.ch)
		}
	case '{': // Left brace for code blocks
		tok = newToken(ast.LBRACE, l.ch)
	case '}': // Right brace for code blocks
		tok = newToken(ast.RBRACE, l.ch)
	case '+': // Addition operator
		tok = newToken(ast.PLUS, l.ch)
	case '-': // Subtraction operator
		tok = newToken(ast.MINUS, l.ch)
	case '*': // Multiplication operator
		tok = newToken(ast.ASTERISK, l.ch)
	case '/': // Division operator
		tok = newToken(ast.SLASH, l.ch)
	case ',': // Comma separator
		tok = newToken(ast.COMMA, l.ch)
	case ';': // Semicolon statement terminator
		tok = newToken(ast.SEMICOLON, l.ch)
	case '(': // Left parenthesis
		tok = newToken(ast.LPAREN, l.ch)
	case ')': // Right parenthesis
		tok = newToken(ast.RPAREN, l.ch)
	case 0: // End of file
		tok.Literal = ""
		tok.Type = ast.EOF
	default:
		if isLetter(l.ch) {
			// Identifier or keyword
			tok.Literal = l.readIdentifier()
			tok.Type = LookupIdent(tok.Literal)
			return &tok
		} else if isDigit(l.ch) {
			// Integer literal
			tok.Type = ast.INT
			tok.Literal = l.readNumber()
			return &tok
		} else {
			// Illegal/unknown character
			tok = newToken(ast.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	return &tok
}

// skipLineComment skips over a line comment (// comment) until end of line.
func (l *Lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	l.skipWhitespace()
}

// skipBlockComment skips over a block comment (/* comment */).
func (l *Lexer) skipBlockComment() {
	l.readChar() // skip '/'
	l.readChar() // skip '*'

	for l.ch != 0 {
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // skip '*'
			l.readChar() // skip '/'
			break
		}
		l.readChar()
	}
	l.skipWhitespace()
}

// peekChar looks at the next character without advancing the lexer.
// Returns 0 if at end of input.
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// readString reads a string literal from the input, including the quotes.
// Handles the entire string until the closing quote.
func (l *Lexer) readString() string {
	l.readChar() // Skip the opening quote
	position := l.position

	for l.ch != '"' && l.ch != 0 {
		l.readChar()
	}

	result := l.input[position:l.position]

	if l.ch == '"' {
		l.readChar() // Skip the closing quote
	}

	return result
}

// newToken creates a new token with the given type and character.
func newToken(tokenType ast.TokenType, ch byte) ast.Token {
	return ast.Token{Type: tokenType, Literal: string(ch)}
}

// readIdentifier reads an identifier from the current position.
// An identifier consists of letters and underscores.
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber reads a integer number from the current position.
// A number consists of digit characters.
func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// skipWhitespace skips over any whitespace characters (space, tab, newline, carriage return).
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// readChar advances the lexer to the next character in the input.
// Sets ch to 0 when the end of input is reached.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

// isLetter checks if a character is a valid letter for identifiers.
// Includes uppercase, lowercase letters, and underscores.
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// isDigit checks if a character is a valid digit (0-9).
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// LookupIdent checks if an identifier is a reserved keyword.
// Returns the appropriate token type for keywords, or IDENT for regular identifiers.
func LookupIdent(ident string) ast.TokenType {
	keywords := map[string]ast.TokenType{
		"print":  ast.PRINT,
		"var":    ast.VAR,
		"if":     ast.IF,
		"extern": ast.EXTERN,
		"func":   ast.FUNC,   // Function definition keyword
		"return": ast.RETURN, // Return statement keyword
	}
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return ast.IDENT
}

// [file content end]
