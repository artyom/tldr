package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func main() {
	log.SetFlags(0)
	flag.Parse()
	if err := run(flag.Args()); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		if st, err := os.Stdin.Stat(); err == nil && st.Mode()&os.ModeCharDevice != 0 {
			return errors.New(usage)
		}
	}
	var llmcliArgs []string
	for _, s := range args {
		llmcliArgs = append(llmcliArgs, "-f", s)
	}
	llmcliArgs = append(llmcliArgs, "What's the tl;dr version of this?")
	cmd := exec.Command("llmcli", llmcliArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if foldbin, err := exec.LookPath("fold"); err == nil {
		foldcmd := exec.Command(foldbin, "-s")
		if foldinput, err := foldcmd.StdinPipe(); err == nil {
			cmd.Stdout = foldinput
			foldcmd.Stdout = os.Stdout
			defer foldcmd.Run()
			defer foldinput.Close() // fold will not exit until standard input is closed
		}
	}
	return cmd.Run()
}

func init() {
	flag.Usage = func() { fmt.Fprintln(flag.CommandLine.Output(), usage) }
}

const usage = "usage: tldr file|url [file|url]"
