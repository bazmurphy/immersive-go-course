package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bazmurphy/immersive-go-course/projects/cli-files/go-cat/cmd"
)

func main() {
	flags := &cmd.Flags{}
	flag.BoolVar(&flags.Number, "n", false, "number all output lines")
	flag.Parse()

	args := flag.Args()

	err := cmd.Execute(flags, args)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}
}
