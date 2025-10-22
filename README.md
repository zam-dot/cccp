# CCCP Compiler

A custom compiler for the CCCP (Custom C-like Compiler Project) programming language that generates C code.

## Project Structure

### Core Compiler Files

#### `ast.go` - Abstract Syntax Tree Definitions
**Purpose**: Defines the data structures that represent the parsed program structure.

**Key Components**:
- **Interfaces**: `Node`, `Statement`, `Expression` - Base interfaces for all AST nodes
- **Token System**: `Token` and `TokenType` definitions for all language tokens
- **Statement Nodes**: 
  - `Program` - Root node containing all statements
  - `LetStatement` - Variable declarations (`var x = 5`)
  - `PrintStatement` - Output statements (`print("hello")`)
  - `IfStatement` - Conditional statements
  - `FunctionStatement` - Named function definitions
  - `ExternStatement` - External function declarations
  - `ReturnStatement` - Function return statements
- **Expression Nodes**:
  - `Identifier` - Variable names
  - `IntegerLiteral` - Numeric values
  - `StringLiteral` - String values
  - `InfixExpression` - Binary operations (`x + y`)
  - `FunctionCall` - Function invocations
  - `FunctionLiteral` - Anonymous functions

#### `lexer.go` - Lexical Analyzer
**Purpose**: Converts source code characters into tokens.

**Key Features**:
- **Token Recognition**: Identifies keywords, identifiers, literals, and operators
- **Whitespace Handling**: Skips spaces, tabs, and newlines
- **Comment Support**: Handles both line (`//`) and block (`/* */`) comments
- **String Literals**: Properly processes quoted strings
- **Multi-character Operators**: Recognizes `==`, `!=`, `...` etc.
- **Error Recovery**: Continues lexing after encountering illegal characters

**Main Functions**:
- `New()` - Creates a new lexer instance
- `NextToken()` - Returns the next token from input
- `readIdentifier()`, `readNumber()`, `readString()` - Token readers

#### `parser.go` - Syntax Parser
**Purpose**: Transforms tokens into an Abstract Syntax Tree using Pratt parsing.

**Key Algorithms**:
- **Pratt Parser**: Handles operator precedence elegantly
- **Recursive Descent**: For statement parsing
- **Expression Parsing**: With configurable precedence levels
- **Error Recovery**: Continues parsing after syntax errors

**Main Components**:
- **Precedence Levels**: `LOWEST`, `EQUALS`, `SUM`, `PRODUCT`, etc.
- **Parsing Tables**: `prefixParseFns` and `infixParseFns` for expression parsing
- **Statement Parsers**: Methods for each statement type (`parseLetStatement`, etc.)
- **Expression Parsers**: Methods for each expression type

**Key Methods**:
- `ParseProgram()` - Main entry point that returns the complete AST
- `parseStatement()` - Dispatcher for statement parsing
- `parseExpression()` - Core Pratt parser for expressions

#### `codegen.go` - Code Generator
**Purpose**: Transforms the AST into equivalent C code.

**Key Features**:
- **Type Inference**: Determines variable types (int vs string)
- **Function Handling**: Generates function declarations and definitions
- **String Support**: Handles string concatenation and comparison
- **Scope Management**: Tracks variables in different scopes
- **Auto-printing**: Automatically prints top-level expressions

**Main Components**:
- **State Management**: Tracks variables, functions, and current context
- **C Code Generation**: Outputs standards-compliant C code
- **Helper Functions**: Generates utility code (string concatenation)

**Key Methods**:
- `Generate()` - Main entry point for code generation
- `generateStatement()` - Dispatcher for statement codegen
- `generateExpression()` - Handles expression code generation

#### `main.go` - Compiler Driver
**Purpose**: Coordinates the entire compilation process and provides user interface.

**Compilation Pipeline**:
1. **File Reading**: Loads source code from file
2. **Lexing**: Converts source to tokens (with debug output)
3. **Parsing**: Builds AST from tokens
4. **AST Display**: Pretty-prints the AST structure
5. **Code Generation**: Converts AST to C code
6. **Compilation**: Invokes GCC to compile generated C code
7. **Execution**: Runs the compiled program

**Key Functions**:
- `main()` - Orchestrates the entire compilation process
- `PrettyPrintAST()` - Displays AST in readable format for debugging

## Compilation Process

1. **Input**: CCCP source code file (e.g., `program.cccp`)
2. **Lexical Analysis**: `lexer.go` converts characters to tokens
3. **Syntax Analysis**: `parser.go` builds AST from tokens
4. **Code Generation**: `codegen.go` converts AST to C code
5. **Compilation**: GCC compiles generated C code to executable
6. **Execution**: The compiled program runs and produces output

## Language Features Supported

- **Variables**: `var x = 5` declarations and assignments
- **Functions**: Named and anonymous functions with parameters
- **Control Flow**: `if` statements with blocks
- **I/O**: `print()` statements for output
- **Operators**: Arithmetic (`+`, `-`, `*`, `/`) and comparison (`==`, `!=`)
- **Types**: Integers and strings with type inference
- **External Functions**: `extern` declarations for C library functions

## Example Usage

```bash
go run main.go examples/hello.cccp
