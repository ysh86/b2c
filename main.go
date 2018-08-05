package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/ysh86/b2c/lexer"
	"github.com/ysh86/b2c/parser"
)

func parse(r io.Reader, w io.Writer) error {
	l := lexer.New(r)
	p := parser.New(l)

	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		p.PrintErrors(w)
		return errors.New("Error!")
	}

	io.WriteString(w, program.String())
	io.WriteString(w, "\n")

	return nil
}

func repl(r io.Reader, w io.Writer) {
	scanner := bufio.NewScanner(r)

	fmt.Println("b2c: a BASIC to C transpiler in golang")
	for {
		fmt.Printf(">> ")

		if ok := scanner.Scan(); !ok {
			return
		}

		line := bytes.NewBufferString(scanner.Text())
		parse(line, w)
	}
}

func main() {
	var isTranspiler bool
	var inFileName string

	flag.BoolVar(&isTranspiler, "c", false, "do transpile")
	flag.Parse()

	if len(os.Args) > 2 {
		inFileName = os.Args[2]
	}

	if isTranspiler {
		file, err := os.Open(inFileName)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		reader := bufio.NewReader(file)

		parse(reader, os.Stdout)
	} else {
		repl(os.Stdin, os.Stdout)
	}
}
