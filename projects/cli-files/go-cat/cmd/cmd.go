package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type Flags struct {
	Number bool
}

func Execute(flags *Flags, args []string) {
	// if there are no arguments provided
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "go-cat: no filename provided\n")
		return
	}

	for _, filename := range args {
		// try to get the file information
		fileInfo, err := os.Stat(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "go-cat: %s: No such file or directory\n", filename)
			continue
		}

		// if the file is a directory then warn the user
		if fileInfo.IsDir() {
			fmt.Fprintf(os.Stderr, "go-cat: %s: Is a directory\n", filename)
			continue
		}

		// try to open the file
		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "go-cat: %v: failed to open the file\n", file)
			continue
		}
		defer file.Close()

		lineNumber := 1

		reader := bufio.NewReader(file)

		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				// if we have reached the end of the file
				if err == io.EOF {
					// if the line is not empty
					if line != "" {
						if flags.Number {
							fmt.Fprintf(os.Stdout, "%d\t%s", lineNumber, line)
							lineNumber++
						} else {
							fmt.Fprint(os.Stdout, line)
						}
					}
					break
				}
				fmt.Fprintf(os.Stderr, "go-cat: %v: failed to read line %d", file, lineNumber)
				continue
			}

			if flags.Number {
				fmt.Fprintf(os.Stdout, "%d\t%s", lineNumber, line)
				lineNumber++
			} else {
				fmt.Fprint(os.Stdout, line)
			}
		}
	}
}
