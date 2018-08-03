package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/ysh86/b2c/lexer"
	"github.com/ysh86/b2c/parser"
)

func repl(r io.Reader, w io.Writer) {
	scanner := bufio.NewScanner(r)

	for {
		fmt.Printf(">> ")

		if ok := scanner.Scan(); !ok {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			p.PrintErrors(w)
			continue
		}

		io.WriteString(w, program.String())
		io.WriteString(w, "\n")
	}
}

func main() {
	fmt.Println("b2c: a BASIC to C transpiler in golang")
	repl(os.Stdin, os.Stdout)
}
