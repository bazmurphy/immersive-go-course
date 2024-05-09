package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bazmurphy/immersive-go-course/projects/cli-files/go-ls/cmd"
)

func main() {
	flags := &cmd.Flags{}
	flag.BoolVar(&flags.Help, "h", false, "show go-ls help")
	flag.Parse()

	args := flag.Args()

	err := cmd.Execute(flags, args)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}
}
