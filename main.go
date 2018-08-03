package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/user"

	"github.com/ysh86/b2c/lexer"
	"github.com/ysh86/b2c/parser"
)

const PROMPT = ">> "

const MONKEY_FACE = `// ***********
// **  B2C  **
// ***********
`

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, MONKEY_FACE)
	io.WriteString(out, "// parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "//  "+msg+"\n")
	}
}

func replStart(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		io.WriteString(out, program.String())
		io.WriteString(out, "\n")
	}
}

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s! This is the Monkey programming language!\n",
		user.Username)
	fmt.Printf("Feel free to type in commands\n")
	replStart(os.Stdin, os.Stdout)
}
