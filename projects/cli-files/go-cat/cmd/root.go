// projects/cli-files/go-cat/cmd/root.go

package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

var (
	numberFlag = flag.Bool("n", false, "number all output lines")
)

func Execute() {
	flag.Parse()

	// get the non-flag arguments
	args := flag.Args()

	// if there is no non-flag argument given then error
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "go-cat: no filename provided\n")
		os.Exit(2)
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
				if err == io.EOF {
					// LAST LINE OF THE FILE
					// ---------- DEBUG

					// fmt.Print(line)

					// actual hellohello | expected hello

					// fmt.Println(line)

					// 	output: actual hello
					//
					//  | expected hello

					// fmt.Printf("%s", line)

					// output: actual hellohello | expected hello

					fmt.Fprint(os.Stdout, line)

					// output: actual hellohello | expected hello

					// fmt.Fprintln(os.Stdout, line)

					// output: actual hello
					//
					//  | expected hello

					// fmt.Fprintf(os.Stdout, "%s", line)

					// output: actual hellohello | expected hello

					break
				}
				fmt.Fprintf(os.Stderr, "go-cat: %v: failed to read line %d", file, lineNumber)
				break
			}

			// REGULAR LINE
			// ---------- DEBUG

			// fmt.Print(line)
			// fmt.Println(line)
			// fmt.Printf("%s", line)

			fmt.Fprint(os.Stdout, line)
			// fmt.Fprintln(os.Stdout, line)
			// fmt.Fprintf(os.Stdout, "%s", line)

			// LINE NUMBERS LOGIC (turn this back on later)
			// if *numberFlag {
			// 	fmt.Fprintf(os.Stdout, "%d\t%s", lineNumber, line)
			// 	lineNumber++
			// } else {
			// 	fmt.Fprintf(os.Stdout, "%s", line)
			// }
		}

	}
}
