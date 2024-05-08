package main

import (
	"fmt"
	"myinterpreter/repl"
	"os"
	"os/user"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %v! This is the Monkey programming language!\n", user.Username)
	fmt.Printf("Feel free to type in commands\n")
	repl.Start(os.Stdin, os.Stdout)
}

//let unless = macro(condition,consequence,alternative){quote(if(!(unquote(condition))){unquote(consequence);}else{unquote(alternative);});};
//unless(10 > 5, puts("not greater"), puts("greater"));
