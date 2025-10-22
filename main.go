// [file name]: main.go
// [file content begin]
package main

import (
	"fmt"
	"os"
	"os/exec"

	"cccp/ast"
	"cccp/codegen"
	"cccp/lexer"
	"cccp/parser"
)

// main is the entry point of the CCCP compiler.
// It handles command-line arguments, coordinates the compilation process,
// and executes the generated code.
func main() {
	// Check if filename was provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <filename>")
		fmt.Println("Example: go run main.go program.cccp")
		return
	}

	filename := os.Args[1]

	// Read source code from file
	inputBytes, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	input := string(inputBytes)

	// Lexical Analysis: Convert source code to tokens
	l := lexer.New(input)

	// DEBUG: Print all tokens to verify lexer output
	fmt.Println("=== Lexer Tokens ===")
	for {
		tok := l.NextToken()
		fmt.Printf("Token: %s '%s'\n", tok.Type, tok.Literal)
		if tok.Type == ast.EOF {
			break
		}
	}

	// Reset lexer for parser (since we consumed tokens during debugging)
	l = lexer.New(input)

	// Syntactic Analysis: Parse tokens into an Abstract Syntax Tree
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parsing errors
	if len(p.Errors()) > 0 {
		fmt.Println("Parser errors:")
		for _, msg := range p.Errors() {
			fmt.Println("\t", msg)
		}
		return
	}

	// Display the parsed AST structure
	fmt.Println("=== AST ===")
	PrettyPrintAST(program, "")

	// DEBUG: Print statement types found
	fmt.Println("\n=== Statements Found ===")
	for i, stmt := range program.Statements {
		fmt.Printf("Statement %d: %T\n", i, stmt)
	}

	// Code Generation: Convert AST to C code
	fmt.Println("\n=== Generated C Code ===")
	generator := codegen.New()
	cCode := generator.Generate(program)
	fmt.Println(cCode)

	// Write generated C code to output file
	err = os.WriteFile("output/output.c", []byte(cCode), 0644)
	if err != nil {
		fmt.Println("Error writing output.c:", err)
		return
	}
	fmt.Println("\n✅ C code written to output.c")

	// Compile the generated C code using GCC
	fmt.Println("\n=== Compiling and Running ===")
	cmd := exec.Command("gcc", "output/output.c", "-o", "output/output")
	compileOutput, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Compilation failed:")
		fmt.Println(string(compileOutput))
		return
	}
	fmt.Println("✅ Compiled successfully!")

	// Execute the compiled program
	runCmd := exec.Command("./output/output")
	runOutput, err := runCmd.Output()
	if err != nil {
		fmt.Println("Execution failed:", err)
		return
	}
	fmt.Printf("✅ Program output: %s", runOutput)
}

// PrettyPrintAST recursively prints the AST in a readable, indented format.
// This is useful for debugging and understanding the structure of the parsed program.
func PrettyPrintAST(node ast.Node, indent string) {
	switch n := node.(type) {
	case *ast.Program:
		fmt.Printf("%sProgram:\n", indent)
		for _, stmt := range n.Statements {
			PrettyPrintAST(stmt, indent+"  ")
		}
	case *ast.LetStatement:
		fmt.Printf("%sLetStatement:\n", indent)
		fmt.Printf("%s  Name: %s\n", indent, n.Name.Value)
		if n.Value != nil {
			fmt.Printf("%s  Value:\n", indent)
			PrettyPrintAST(n.Value, indent+"    ")
		}
	case *ast.PrintStatement:
		fmt.Printf("%sPrintStatement:\n", indent)
		if n.Value != nil {
			fmt.Printf("%s  Value:\n", indent)
			PrettyPrintAST(n.Value, indent+"    ")
		}
	case *ast.ExternStatement:
		fmt.Printf("%sExternStatement:\n", indent)
		fmt.Printf("%s  Name: %s\n", indent, n.Name.Value)
	case *ast.Identifier:
		fmt.Printf("%sIdentifier: %s\n", indent, n.Value)
	case *ast.IntegerLiteral:
		fmt.Printf("%sInteger: %d\n", indent, n.Value)
	case *ast.StringLiteral:
		fmt.Printf("%sString: %s\n", indent, n.Value)
	case *ast.InfixExpression:
		fmt.Printf("%sInfixExpression: (%s)\n", indent, n.Operator)
		fmt.Printf("%s  Left:\n", indent)
		PrettyPrintAST(n.Left, indent+"    ")
		fmt.Printf("%s  Right:\n", indent)
		PrettyPrintAST(n.Right, indent+"    ")
	case *ast.FunctionCall:
		fmt.Printf("%sFunctionCall: %s\n", indent, n.Token.Literal)
		fmt.Printf("%s  Function:\n", indent)
		PrettyPrintAST(n.Function, indent+"    ")
		if len(n.Arguments) > 0 {
			fmt.Printf("%s  Arguments:\n", indent)
			for _, arg := range n.Arguments {
				PrettyPrintAST(arg, indent+"    ")
			}
		}
	case *ast.ExpressionStatement:
		fmt.Printf("%sExpressionStatement:\n", indent)
		if n.Expression != nil {
			PrettyPrintAST(n.Expression, indent+"  ")
		}
	}
}

// [file content end]
