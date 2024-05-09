package main

import (
	"flag"

	"github.com/bazmurphy/immersive-go-course/projects/cli-files/go-cat/cmd"
)

func main() {
	flags := &cmd.Flags{}
	flag.BoolVar(&flags.Number, "n", false, "number all output lines")
	flag.Parse()

	args := flag.Args()

	cmd.Execute(flags, args)
}
