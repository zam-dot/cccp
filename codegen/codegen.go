// [file name]: codegen.go
// [file content begin]
package codegen

import (
	"cccp/ast"
	"fmt"
	"maps"
	"strings"
)

// CodeGenerator translates the AST into equivalent C code.
// It maintains state about variables, functions, and the current output.
type CodeGenerator struct {
	output      strings.Builder                 // Accumulates the generated C code
	variables   map[string]string               // Tracks variable names and their types
	functions   map[string]*ast.FunctionLiteral // Stores function definitions
	currentFunc *ast.FunctionLiteral            // Tracks the function being generated
	funcCounter int                             // Counter for generating unique names for anonymous functions
}

// New creates and initializes a new CodeGenerator.
func New() *CodeGenerator {
	return &CodeGenerator{
		variables:   make(map[string]string),
		functions:   make(map[string]*ast.FunctionLiteral),
		funcCounter: 0,
	}
}

// Generate converts the entire program AST into C source code.
// It performs multiple passes: function extraction, declarations, definitions, and main generation.
func (cg *CodeGenerator) Generate(program *ast.Program) string {
	// Reset generator state for a fresh compilation
	cg.output.Reset()
	cg.variables = make(map[string]string)
	cg.functions = make(map[string]*ast.FunctionLiteral)
	cg.currentFunc = nil
	cg.funcCounter = 0

	// Write C standard library includes
	cg.output.WriteString("#include <stdio.h>\n")
	cg.output.WriteString("#include <string.h>\n")
	cg.output.WriteString("#include <stdlib.h>\n\n")

	// Generate helper function for string concatenation
	cg.generateStringConcatHelper()

	// First pass: extract function definitions from the program
	mainStatements := cg.extractFunctions(program.Statements)

	// Generate forward declarations for all functions
	cg.generateFunctionDeclarations()

	// Generate definitions for all functions
	cg.generateFunctionDefinitions()

	// Generate the main function with remaining statements
	cg.generateMainFunction(mainStatements)

	return cg.output.String()
}

// generateStringConcatHelper generates a C helper function for concatenating strings.
func (cg *CodeGenerator) generateStringConcatHelper() {
	cg.output.WriteString("char* concat_strings(const char* a, const char* b) {\n")
	cg.output.WriteString("    char* result = malloc(strlen(a) + strlen(b) + 1);\n")
	cg.output.WriteString("    strcpy(result, a);\n")
	cg.output.WriteString("    strcat(result, b);\n")
	cg.output.WriteString("    return result;\n")
	cg.output.WriteString("}\n\n")
}

// extractFunctions processes statements, extracting function definitions and
// returning non-function statements for the main program.
func (cg *CodeGenerator) extractFunctions(statements []ast.Statement) []ast.Statement {
	mainStatements := []ast.Statement{}

	for _, stmt := range statements {
		switch s := stmt.(type) {
		case *ast.FunctionStatement:
			// Register named function definition
			funcName := s.Name.Value
			funcLit := &ast.FunctionLiteral{
				Token:      s.Token,
				Parameters: s.Parameters,
				Body:       s.Body,
			}
			cg.functions[funcName] = funcLit
		case *ast.ExpressionStatement:
			// Handle anonymous function expressions
			if funcLit, ok := s.Expression.(*ast.FunctionLiteral); ok {
				funcName := fmt.Sprintf("func_%d", cg.funcCounter)
				cg.funcCounter++
				cg.functions[funcName] = funcLit
			} else {
				// Add non-function expression statements to main
				mainStatements = append(mainStatements, stmt)
			}
		default:
			// Keep all other statement types for main program
			mainStatements = append(mainStatements, stmt)
		}
	}

	return mainStatements
}

// generateFunctionDeclarations generates forward declarations for all functions.
func (cg *CodeGenerator) generateFunctionDeclarations() {
	for funcName := range cg.functions {
		cg.generateFunctionDeclaration(funcName)
	}
	cg.output.WriteString("\n")
}

// generateFunctionDeclaration generates a single function forward declaration.
func (cg *CodeGenerator) generateFunctionDeclaration(funcName string) {
	funcLit := cg.functions[funcName]
	cg.output.WriteString("int ")
	cg.output.WriteString(funcName)
	cg.output.WriteString("(")

	// Generate parameter list
	for i, param := range funcLit.Parameters {
		if i > 0 {
			cg.output.WriteString(", ")
		}
		cg.output.WriteString("int ")
		cg.output.WriteString(param.Value)
	}
	cg.output.WriteString(");\n")
}

// generateFunctionDefinitions generates complete definitions for all functions.
func (cg *CodeGenerator) generateFunctionDefinitions() {
	for funcName, funcLit := range cg.functions {
		cg.generateFunctionDefinition(funcName, funcLit)
		cg.output.WriteString("\n")
	}
}

// generateFunctionDefinition generates a complete function definition in C.
func (cg *CodeGenerator) generateFunctionDefinition(funcName string, funcLit *ast.FunctionLiteral) {
	// Save current generator state
	oldVariables := cg.backupVariables()
	oldFunc := cg.currentFunc

	// Setup new function context
	cg.variables = make(map[string]string)
	cg.currentFunc = funcLit

	// Generate function header
	cg.generateFunctionHeader(funcName, funcLit)

	// Generate function body
	cg.generateFunctionBody(funcName, funcLit)

	// Add default return if needed
	cg.addDefaultReturnIfNeeded(funcName, funcLit)

	cg.output.WriteString("}\n")

	// Restore previous generator state
	cg.restoreGeneratorState(oldVariables, oldFunc)
}

// generateFunctionHeader generates the function signature in C.
func (cg *CodeGenerator) generateFunctionHeader(funcName string, funcLit *ast.FunctionLiteral) {
	cg.output.WriteString("int ")
	cg.output.WriteString(funcName)
	cg.output.WriteString("(")

	// Process parameters
	for i, param := range funcLit.Parameters {
		if i > 0 {
			cg.output.WriteString(", ")
		}
		cg.output.WriteString("int ")
		cg.output.WriteString(param.Value)
		// Register parameters as variables in current scope
		cg.variables[param.Value] = "int"
		fmt.Printf(
			"🔍 DEBUG: Registered parameter %s as int in function %s\n",
			param.Value,
			funcName,
		)
	}
	cg.output.WriteString(") {\n")
}

// generateFunctionBody generates the statements within a function body.
func (cg *CodeGenerator) generateFunctionBody(funcName string, funcLit *ast.FunctionLiteral) {
	fmt.Printf("🔍 DEBUG: Variables in function %s scope: %v\n", funcName, cg.variables)

	for _, stmt := range funcLit.Body.Statements {
		cg.generateStatement(stmt)
	}
}

// addDefaultReturnIfNeeded adds a default return statement if the function doesn't have one.
func (cg *CodeGenerator) addDefaultReturnIfNeeded(funcName string, funcLit *ast.FunctionLiteral) {
	hasExplicitReturn := cg.hasExplicitReturn(funcLit)

	if !hasExplicitReturn {
		fmt.Printf("🔍 DEBUG: Adding default return 0 to function %s\n", funcName)
		cg.output.WriteString("    return 0;\n")
	} else {
		fmt.Printf("🔍 DEBUG: Function %s has explicit return, no default needed\n", funcName)
	}
}

// hasExplicitReturn checks if the function ends with an explicit return statement.
func (cg *CodeGenerator) hasExplicitReturn(funcLit *ast.FunctionLiteral) bool {
	if len(funcLit.Body.Statements) > 0 {
		_, hasExplicitReturn := funcLit.Body.Statements[len(funcLit.Body.Statements)-1].(*ast.ReturnStatement)
		return hasExplicitReturn
	}
	return false
}

// generateMainFunction generates the main function with the provided statements.
func (cg *CodeGenerator) generateMainFunction(statements []ast.Statement) {
	cg.output.WriteString("int main() {\n")

	// Generate main function body
	for _, stmt := range statements {
		cg.generateStatement(stmt)
	}

	cg.output.WriteString("    return 0;\n")
	cg.output.WriteString("}\n")
}

// backupVariables creates a copy of the current variables map.
func (cg *CodeGenerator) backupVariables() map[string]string {
	oldVariables := make(map[string]string)
	maps.Copy(oldVariables, cg.variables)
	return oldVariables
}

// restoreGeneratorState restores the generator's previous state after function generation.
func (cg *CodeGenerator) restoreGeneratorState(
	oldVariables map[string]string,
	oldFunc *ast.FunctionLiteral,
) {
	cg.variables = oldVariables
	cg.currentFunc = oldFunc
}

// generateStatement dispatches to the appropriate statement generation method.
func (cg *CodeGenerator) generateStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		cg.generateLetStatement(s)
	case *ast.PrintStatement:
		cg.generatePrintStatement(s)
	case *ast.AssignmentStatement:
		cg.generateAssignmentStatement(s)
	case *ast.IfStatement:
		cg.generateIfStatement(s)
	case *ast.BlockStatement:
		cg.generateBlockStatement(s)
	case *ast.ExternStatement:
		cg.generateExternStatement(s)
	case *ast.ReturnStatement:
		cg.generateReturnStatement(s)
	case *ast.FunctionStatement:
		cg.generateFunctionStatement(s)
	case *ast.ExpressionStatement:
		cg.generateAutoPrint(s.Expression)
	}
}

// generateAutoPrint automatically generates print calls for top-level expressions.
// Inside functions, expressions are evaluated without printing.
func (cg *CodeGenerator) generateAutoPrint(exp ast.Expression) {
	// Only auto-print in main function, not inside other functions
	if cg.currentFunc == nil {
		cg.output.WriteString("    // Auto-print: ")
		cg.generateExpression(exp)
		cg.output.WriteString("\n")

		// Generate appropriate printf call based on expression type
		switch exp.(type) {
		case *ast.StringLiteral:
			cg.output.WriteString(`    printf("%s\n", `)
			cg.generateExpression(exp)
			cg.output.WriteString(");\n")
		default:
			cg.output.WriteString(`    printf("%d\n", `)
			cg.generateExpression(exp)
			cg.output.WriteString(");\n")
		}
	} else {
		// Inside functions, just evaluate the expression
		cg.generateExpression(exp)
		cg.output.WriteString(";\n")
	}
}

// generateFunctionStatement registers a function statement for later code generation.
func (cg *CodeGenerator) generateFunctionStatement(stmt *ast.FunctionStatement) {
	funcName := stmt.Name.Value
	funcLit := &ast.FunctionLiteral{
		Token:      stmt.Token,
		Parameters: stmt.Parameters,
		Body:       stmt.Body,
	}
	cg.functions[funcName] = funcLit
}

// generateReturnStatement generates a return statement in C.
func (cg *CodeGenerator) generateReturnStatement(stmt *ast.ReturnStatement) {
	cg.output.WriteString("    return")
	if stmt.ReturnValue != nil {
		cg.output.WriteString(" ")
		cg.generateExpression(stmt.ReturnValue)
	}
	cg.output.WriteString(";\n")
}

// generateFunctionCall generates a function call expression in C.
func (cg *CodeGenerator) generateFunctionCall(call *ast.FunctionCall) {
	if ident, ok := call.Function.(*ast.Identifier); ok {
		// Use actual function name for defined functions, fall back to identifier for externals
		if _, exists := cg.functions[ident.Value]; exists {
			cg.output.WriteString(ident.Value)
		} else {
			cg.output.WriteString(ident.Value)
		}
	} else {
		cg.generateExpression(call.Function)
	}

	cg.output.WriteString("(")
	// Generate argument list
	for i, arg := range call.Arguments {
		if i > 0 {
			cg.output.WriteString(", ")
		}
		cg.generateExpression(arg)
	}
	cg.output.WriteString(")")
}

// generateLetStatement generates a variable declaration and initialization in C.
func (cg *CodeGenerator) generateLetStatement(stmt *ast.LetStatement) {
	varName := stmt.Name.Value
	fmt.Printf("🔍 DEBUG generateLetStatement: %s = %T\n", varName, stmt.Value)

	if stmt.Value != nil {
		// Determine variable type based on the expression
		if cg.expressionIsString(stmt.Value) {
			cg.variables[varName] = "string"
			cg.output.WriteString("    char* ")
			cg.output.WriteString(varName)
			cg.output.WriteString(" = ")
			fmt.Printf("🔍 DEBUG: Generating string expression for %s\n", varName)
			cg.generateExpression(stmt.Value)
			cg.output.WriteString(";\n")
		} else {
			cg.variables[varName] = "int"
			cg.output.WriteString("    int ")
			cg.output.WriteString(varName)
			cg.output.WriteString(" = ")
			fmt.Printf("🔍 DEBUG: Generating int expression for %s: %T\n", varName, stmt.Value)
			cg.generateExpression(stmt.Value)
			cg.output.WriteString(";\n")
		}
	} else {
		// Variable declaration without initialization
		cg.variables[varName] = "int"
		cg.output.WriteString("    int ")
		cg.output.WriteString(varName)
		cg.output.WriteString(";\n")
	}
}

// expressionIsString determines if an expression evaluates to a string type.
func (cg *CodeGenerator) expressionIsString(exp ast.Expression) bool {
	switch e := exp.(type) {
	case *ast.StringLiteral:
		return true
	case *ast.InfixExpression:
		// Check for string concatenation
		if e.Operator == "+" && cg.isSimpleStringConcat(e) {
			return true
		}
		return false
	case *ast.Identifier:
		// Check variable type from symbol table
		if varType, exists := cg.variables[e.Value]; exists {
			return varType == "string"
		}
		return false
	default:
		return false
	}
}

// generateAssignmentStatement generates a variable assignment in C.
// Declares the variable if it hasn't been declared yet.
func (cg *CodeGenerator) generateAssignmentStatement(stmt *ast.AssignmentStatement) {
	varName := stmt.Name.Value

	// Check if this variable hasn't been declared yet
	if _, exists := cg.variables[varName]; !exists {
		// Variable needs to be declared
		if cg.expressionIsString(stmt.Value) {
			cg.variables[varName] = "string"
			cg.output.WriteString("    char* ")
		} else {
			cg.variables[varName] = "int"
			cg.output.WriteString("    int ")
		}
		cg.output.WriteString(varName)
		cg.output.WriteString(" = ")
	} else {
		// Variable already declared, just assignment
		cg.output.WriteString("    ")
		cg.output.WriteString(varName)
		cg.output.WriteString(" = ")
	}

	cg.generateExpression(stmt.Value)
	cg.output.WriteString(";\n")
}

// generatePrintStatement generates a printf call appropriate for the value type.
func (cg *CodeGenerator) generatePrintStatement(stmt *ast.PrintStatement) {
	switch expr := stmt.Value.(type) {
	case *ast.Identifier:
		if varType, exists := cg.variables[expr.Value]; exists && varType == "string" {
			cg.output.WriteString(`    printf("%s\n", `)
			cg.generateExpression(stmt.Value)
			cg.output.WriteString(");\n")
		} else {
			cg.output.WriteString(`    printf("%d\n", `)
			cg.generateExpression(stmt.Value)
			cg.output.WriteString(");\n")
		}
	case *ast.StringLiteral:
		cg.output.WriteString(`    printf("%s\n", `)
		cg.generateExpression(stmt.Value)
		cg.output.WriteString(");\n")
	default:
		cg.output.WriteString(`    printf("%d\n", `)
		cg.generateExpression(stmt.Value)
		cg.output.WriteString(");\n")
	}
}

// generateIfStatement generates an if statement in C with proper scoping.
func (cg *CodeGenerator) generateIfStatement(stmt *ast.IfStatement) {
	// Save current variable state for scoping
	oldVariables := cg.backupVariables()

	cg.output.WriteString("    if (")
	cg.generateExpression(stmt.Condition)
	cg.output.WriteString(") {\n")
	cg.generateBlockStatement(stmt.Consequence)
	cg.output.WriteString("    }\n")

	// Restore variable state after the block
	cg.variables = oldVariables
}

// generateBlockStatement generates a block of statements with proper variable scoping.
func (cg *CodeGenerator) generateBlockStatement(block *ast.BlockStatement) {
	// Save current variable state at block entry
	blockVariables := cg.backupVariables()

	// Generate all statements in the block
	for _, stmt := range block.Statements {
		cg.generateStatement(stmt)
	}

	// Restore variable state after block
	cg.variables = blockVariables
}

// generateExternStatement generates a comment for extern declarations.
// External functions are assumed to be provided by C headers.
func (cg *CodeGenerator) generateExternStatement(stmt *ast.ExternStatement) {
	cg.output.WriteString("    // extern ")
	cg.output.WriteString(stmt.Name.Value)
	cg.output.WriteString(" declared (handled by C headers)\n")
}

// generateExpression dispatches to the appropriate expression generation method.
func (cg *CodeGenerator) generateExpression(exp ast.Expression) {
	fmt.Printf("🔍 DEBUG generateExpression: %T\n", exp)
	switch e := exp.(type) {
	case *ast.Identifier:
		fmt.Printf("🔍 DEBUG: Generating identifier: %s (known variables: %v)\n", e.Value, cg.variables)
		cg.output.WriteString(e.Value)
	case *ast.IntegerLiteral:
		fmt.Printf("🔍 DEBUG: Generating integer: %d\n", e.Value)
		cg.output.WriteString(fmt.Sprintf("%d", e.Value))
	case *ast.StringLiteral:
		fmt.Printf("🔍 DEBUG: Generating string: %s\n", e.Value)
		cg.output.WriteString(`"`)
		cg.output.WriteString(e.Value)
		cg.output.WriteString(`"`)
	case *ast.InfixExpression:
		fmt.Printf("🔍 DEBUG: Generating infix expression: %s\n", e.Operator)
		if e.Operator == "+" && cg.isSimpleStringConcat(e) {
			cg.generateSimpleStringConcat(e)
		} else if (e.Operator == "==" || e.Operator == "!=") && cg.isStringComparison(e) {
			cg.generateStringComparison(e)
		} else {
			// Standard arithmetic or comparison operation
			cg.generateExpression(e.Left)
			cg.output.WriteString(" ")
			cg.output.WriteString(e.Operator)
			cg.output.WriteString(" ")
			cg.generateExpression(e.Right)
		}
	case *ast.FunctionCall:
		fmt.Printf("🔍 DEBUG: Generating function call\n")
		cg.generateFunctionCall(e)
	default:
		fmt.Printf("❌ DEBUG: Unknown expression type in generateExpression: %T\n", exp)
	}
}

// isStringComparison checks if an infix expression is comparing strings.
func (cg *CodeGenerator) isStringComparison(exp *ast.InfixExpression) bool {
	return cg.expressionIsString(exp.Left) || cg.expressionIsString(exp.Right)
}

// generateStringComparison generates string comparison using strcmp.
func (cg *CodeGenerator) generateStringComparison(exp *ast.InfixExpression) {
	switch exp.Operator {
	case "==":
		cg.output.WriteString("(strcmp(")
		cg.generateExpression(exp.Left)
		cg.output.WriteString(", ")
		cg.generateExpression(exp.Right)
		cg.output.WriteString(") == 0)")
	case "!=":
		cg.output.WriteString("(strcmp(")
		cg.generateExpression(exp.Left)
		cg.output.WriteString(", ")
		cg.generateExpression(exp.Right)
		cg.output.WriteString(") != 0)")
	}
	// No default case needed since we only call this for == and != operators
}

// isSimpleStringConcat checks if an infix expression is a simple string concatenation.
func (cg *CodeGenerator) isSimpleStringConcat(exp *ast.InfixExpression) bool {
	if exp.Operator != "+" {
		return false
	}

	// Check if operands are string literals or string variables
	_, leftIsString := exp.Left.(*ast.StringLiteral)
	_, rightIsString := exp.Right.(*ast.StringLiteral)

	leftIsStringVar := cg.isStringVariable(exp.Left)
	rightIsStringVar := cg.isStringVariable(exp.Right)

	// Handle: "string" + "string", "string" + variable, variable + "string"
	return (leftIsString && rightIsString) ||
		(leftIsString && rightIsStringVar) ||
		(leftIsStringVar && rightIsString)
}

// isStringVariable checks if an expression is a variable of string type.
func (cg *CodeGenerator) isStringVariable(exp ast.Expression) bool {
	if ident, ok := exp.(*ast.Identifier); ok {
		if varType, exists := cg.variables[ident.Value]; exists && varType == "string" {
			return true
		}
	}
	return false
}

// generateSimpleStringConcat generates string concatenation using the helper function.
func (cg *CodeGenerator) generateSimpleStringConcat(exp *ast.InfixExpression) {
	leftStr, leftIsString := exp.Left.(*ast.StringLiteral)
	rightStr, rightIsString := exp.Right.(*ast.StringLiteral)

	leftIdent, leftIsIdent := exp.Left.(*ast.Identifier)
	rightIdent, rightIsIdent := exp.Right.(*ast.Identifier)

	// Case 1: "string" + "string" - compile-time concatenation
	if leftIsString && rightIsString {
		cg.output.WriteString(`"`)
		cg.output.WriteString(leftStr.Value)
		cg.output.WriteString(rightStr.Value)
		cg.output.WriteString(`"`)
		return
	}

	// Case 2: "string" + variable - runtime concatenation
	if leftIsString && rightIsIdent {
		cg.output.WriteString(`concat_strings("`)
		cg.output.WriteString(leftStr.Value)
		cg.output.WriteString(`", `)
		cg.output.WriteString(rightIdent.Value)
		cg.output.WriteString(`)`)
		return
	}

	// Case 3: variable + "string" - runtime concatenation
	if leftIsIdent && rightIsString {
		cg.output.WriteString(`concat_strings(`)
		cg.output.WriteString(leftIdent.Value)
		cg.output.WriteString(`, "`)
		cg.output.WriteString(rightStr.Value)
		cg.output.WriteString(`")`)
		return
	}

	// Fallback for unsupported cases
	cg.output.WriteString(`"concat_error"`)
}

// [file content end]
