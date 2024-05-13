package compiler

import (
	"fmt"
	"myinterpreter/ast"
	"myinterpreter/code"
	"myinterpreter/lexer"
	"myinterpreter/object"
	"myinterpreter/parser"
	"testing"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []any
	expectedInstructions []code.Instructions
}

func TestConditionals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
if(true) {10};3333;
`, //if(true){10}else{Null};3333;
			expectedConstants: []any{10, 3333},
			expectedInstructions: []code.Instructions{
				// 0000
				code.Make(code.OpTrue),
				// 0001
				code.Make(code.OpJumpNotTruthy, 10),
				// 0004
				code.Make(code.OpConstant, 0),
				// 0007
				code.Make(code.OpJump, 11),
				// 0010
				code.Make(code.OpNull),
				// 0011
				code.Make(code.OpPop),
				// 0012
				code.Make(code.OpConstant, 1),
				// 0015
				code.Make(code.OpPop),
			},
		},
		{
			input: `
if (true) { 10 } else { 20 }; 3333;
`,
			expectedConstants: []interface{}{10, 20, 3333}, expectedInstructions: []code.Instructions{
				// 0000
				code.Make(code.OpTrue),
				// 0001
				code.Make(code.OpJumpNotTruthy, 10),
				// 0004
				code.Make(code.OpConstant, 0),
				// 0007
				code.Make(code.OpJump, 13),
				// 0010
				code.Make(code.OpConstant, 1),
				// 0013
				code.Make(code.OpPop),
				// 0014
				code.Make(code.OpConstant, 2),
				// 0017
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1 + 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1;2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: "1 - 2", expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSub),
				code.Make(code.OpPop),
			},
		},
		{
			input: "1 * 2", expectedConstants: []interface{}{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpMul),
				code.Make(code.OpPop),
			},
		},
		{
			input: "2 / 1", expectedConstants: []interface{}{2, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpDiv),
				code.Make(code.OpPop),
			},
		},
		{
			input: "-1", expectedConstants: []interface{}{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpMinus),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()

	for _, ts := range tests {
		program := parse(ts.input)
		compiler := New()
		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}
		byteCode := compiler.Bytecode()
		err = testInstructions(ts.expectedInstructions, byteCode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions failed: %s", err)
		}
		err = testConstants(t, ts.expectedConstants, byteCode.Constants)
		if err != nil {
			t.Fatalf("testConstants failed:%s", err)
		}
	}
}

func parse(input string) *ast.Program {
	l := lexer.New(input)
	p := parser.New(l)
	return p.ParseProgram()
}

func concatInstruction(instructions []code.Instructions) code.Instructions {
	out := code.Instructions{}

	for _, v := range instructions {
		out = append(out, v...)
	}
	return out
}

func testInstructions(expected []code.Instructions, actual code.Instructions) error {
	concatedInstructions := concatInstruction(expected)
	if len(concatedInstructions) != len(actual) {

		return fmt.Errorf("wrong instructions length. \nwant=%q,\ngot=%q", concatedInstructions, actual)
	}
	for i, ins := range concatedInstructions {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction at %d. \nwant=%q,\ngot=%q", i, concatedInstructions, actual)
		}
	}
	return nil
}

func testConstants(t *testing.T, expected []any, actual []object.Object) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. want=%d,got=%q", len(expected), len(actual))
	}
	for i, constant := range expected {
		switch constType := constant.(type) {
		case int:
			err := testIntegerObject(actual[i], int64(constType))
			if err != nil {
				return fmt.Errorf("constant %d -- testIntegerObject failed:%s", i, err.Error())
			}
		case string:
			err := testStringObject(actual[i], constType)
			if err != nil {
				return fmt.Errorf("constant %d - testStringObject failed:%s", i, err)
			}
		case []code.Instructions:
			fn, ok := actual[i].(*object.CompiledFunction)
			if !ok {
				return fmt.Errorf("constant %d - not a function: %T", i, actual[i])
			}

			err := testInstructions(constType, fn.Instructions)
			if err != nil {
				return fmt.Errorf("constant %d - testInstructions failed: %s", i, err)
			}
		}
	}
	return nil
}

func testStringObject(actual object.Object, expected string) error {
	res, ok := actual.(*object.String)
	if !ok {
		return fmt.Errorf("object is not string.got=%T(%+v)", actual, actual)
	}
	if res.Value != expected {
		return fmt.Errorf("object has wrong value. got=%q,want=%q", res.Value, expected)
	}
	return nil
}

func testIntegerObject(obj object.Object, expected int64) error {
	result, ok := obj.(*object.Integer)
	if !ok {
		return fmt.Errorf("object is not Int. got=%T(%+v)", obj, obj)
	}
	if result.Value != expected {
		return fmt.Errorf("object has wrong value. got=%d,want=%d", result.Value, expected)
	}
	return nil
}

func TestBooleanExpression(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "true",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "false",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1>2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1<2",
			expectedConstants: []any{2, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpGreaterThan),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1==2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1!=2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true==false",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true!=false",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpFalse),
				code.Make(code.OpNotEqual),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "!true",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpBang),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestGlobalLetStatements(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
			let one=1;
			let two=2;
`,
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 1),
			},
		},
		{
			input: `
			let one=1;
			one;
`,
			expectedConstants: []any{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
			let one=1;
			let two=one;
			two;
`,
			expectedConstants: []any{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestStringExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `"monkey"`,
			expectedConstants: []any{"monkey"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `"mon"+"key"`,
			expectedConstants: []any{"mon", "key"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestArrayLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[]",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpArray, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[1,2,3]",
			expectedConstants: []any{1, 2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[1+2,3-4,5*6]",
			expectedConstants: []any{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSub),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMul),
				code.Make(code.OpArray, 3),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestHashLiteral(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "{}",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpHash, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1: 2, 3: 4, 5: 6}",
			expectedConstants: []any{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpHash, 6),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1: 2+3, 4: 5*6}",
			expectedConstants: []any{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpAdd),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMul),
				code.Make(code.OpHash, 4),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestIndexExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[1, 2, 3][1 + 1]",
			expectedConstants: []interface{}{1, 2, 3, 1, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpArray, 3),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpAdd),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1:2}[2-1]",
			expectedConstants: []interface{}{1, 2, 2, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpHash, 2),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpSub),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestFunctions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `fn(){return 5+10}`,
			expectedConstants: []any{5, 10, []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpReturnValue),
			}},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn(){5+10}`,
			expectedConstants: []any{5, 10, []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpReturnValue),
			}},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn() { 1; 2 }`,
			expectedConstants: []any{
				1,
				2,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpPop),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 2), code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestCompilerScopes(t *testing.T) {
	compiler := New()
	if compiler.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", compiler.scopeIndex, 0)
	}
	globalSymbolTable := compiler.symbolTable
	compiler.emit(code.OpMul)

	compiler.enterScope()
	if compiler.scopeIndex != 1 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", compiler.scopeIndex, 1)
	}
	compiler.emit(code.OpSub)
	if len(compiler.scopes[compiler.scopeIndex].instructions) != 1 {
		t.Errorf("instructions length wrong. got=%d",
			len(compiler.scopes[compiler.scopeIndex].instructions))
	}

	last := compiler.scopes[compiler.scopeIndex].lastInstruction
	if last.Opcode != code.OpSub {
		t.Errorf("lastInstruction.Opcode wrong. got=%d, want=%d", last.Opcode, code.OpSub)
	}
	if compiler.symbolTable.Outer != globalSymbolTable {
		t.Errorf("compiler did not enclose symbolTable")
	}

	compiler.leaveScope()
	if compiler.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d",
			compiler.scopeIndex, 0)
	}
	if compiler.symbolTable != globalSymbolTable {
		t.Errorf("compiler did not restore global symbol table")
	}
	if compiler.symbolTable.Outer != nil {
		t.Errorf("compiler modified global symbol table incorrectly")
	}

	compiler.emit(code.OpAdd)
	if len(compiler.scopes[compiler.scopeIndex].instructions) != 2 {
		t.Errorf("instructions length wrong. got=%d",
			len(compiler.scopes[compiler.scopeIndex].instructions))
	}

	last = compiler.scopes[compiler.scopeIndex].lastInstruction
	if last.Opcode != code.OpAdd {
		t.Errorf("lastInstruction.Opcode wrong. got=%d, want=%d", last.Opcode, code.OpAdd)
	}

	prev := compiler.scopes[compiler.scopeIndex].previousInstruction
	if prev.Opcode != code.OpMul {
		t.Errorf("previousInstruction.Opcode wrong. got=%d, want=%d", prev.Opcode, code.OpMul)
	}
}

func TestFunctionWithoutReturnValue(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `fn() {}`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestFunctionCalls(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `fn() {24}();`,
			expectedConstants: []any{
				24,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 1),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
let noArg = fn() {24};
noArg();
`,
			expectedConstants: []any{
				24,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
let oneArg = fn(a) {};
oneArg(24);
`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
				24,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
		let manyArg = fn(a,b,c) {};
		manyArg(11,22,33);
		`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
				11,
				22,
				33,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpCall, 3),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
let oneArg = fn(a) {a};
oneArg(24);
`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
				},
				24,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
		let manyArg = fn(a,b,c) {a;b;c;};
		manyArg(11,22,33);
		`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 2),
					code.Make(code.OpReturnValue),
				},
				11,
				22,
				33,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpCall, 3),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestLetStatementScope(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
let num = 55;
fn() {num}
`,
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
			expectedConstants: []any{
				55,
				[]code.Instructions{
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpReturnValue),
				},
			},
		},
		{
			input: `
fn() {
	let num = 55;
	num
}
`,
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
			expectedConstants: []any{
				55,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
				},
			},
		},
		{
			input: `
fn() {
	let a = 55;
	let b = 77;
	a+b
}
`,
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 2),
				code.Make(code.OpPop),
			},
			expectedConstants: []any{
				55,
				77,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpSetLocal, 1),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
			},
		},
	}
	runCompilerTests(t, tests)
}
