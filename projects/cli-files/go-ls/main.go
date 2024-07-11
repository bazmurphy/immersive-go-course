package main

import (
	"flag"

	"github.com/bazmurphy/immersive-go-course/projects/cli-files/go-ls/cmd"
)

func main() {
	flags := &cmd.Flags{}
	flag.BoolVar(&flags.Help, "h", false, "show go-ls help")
	flag.BoolVar(&flags.All, "a", false, "show all files (including hidden files)")
	flag.Parse()

	args := flag.Args()

	cmd.Execute(flags, args)
}
