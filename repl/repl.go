package repl

import (
	"bufio"
	"fmt"
	"io"
	"myinterpreter/compiler"
	"myinterpreter/lexer"
	"myinterpreter/parser"
	"myinterpreter/vm"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	//env := object.NewEnvironment()
	//macroEnv := object.NewEnvironment()

	for {
		fmt.Fprintf(out, PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParseErrors(out, p.Errors())
			continue
		}
		//evaluator.Definemacros(program, macroEnv)
		//expanded := evaluator.ExpandMacro(program, macroEnv)
		//eval := evaluator.Eval(expanded, env)
		//if eval != nil {
		//	io.WriteString(out, eval.Inspect())
		//	io.WriteString(out, "\n")
		//}

		compile := compiler.New()
		err := compile.Compile(program)
		if err != nil {
			fmt.Fprintf(out, "Compilation failed:\n %s\n", err)
			continue
		}
		ma := vm.New(compile.Bytecode())
		err = ma.Run()
		if err != nil {
			fmt.Fprintf(out, "Executing bytecode failed:\n %s\n", err)
			continue
		}
		stackTop := ma.StackTop()
		io.WriteString(out, stackTop.Inspect())
		io.WriteString(out, "\n")
	}
}

func printParseErrors(out io.Writer, errors []string) {
	io.WriteString(out, "\t parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
