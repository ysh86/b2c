package main

import (
	"bufio"
	"bytes"
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

	p.ParseProgram(func(s string, isErrors bool) {
		if s != "" {
			io.WriteString(w, s)
			io.WriteString(w, "\n")
		}
		if isErrors {
			p.PrintErrors(w)
		}
	})

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
