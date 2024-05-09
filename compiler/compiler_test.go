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

func TestIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1 + 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
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
		}
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
