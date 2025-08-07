package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
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
	for _, env := range [...]string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if s, ok := os.LookupEnv(env); ok && s != "" {
			// "en_US.CODESET@modifier" into "en_US"
			s, _, _ = strings.Cut(s, ".")
			s, _, _ = strings.Cut(s, "@")
			if tag, err := language.Parse(s); err == nil {
				llmcliArgs = append(llmcliArgs, "Please respond in "+display.English.Languages().Name(tag)+".")
				break
			}
		}
	}
	cmd := exec.Command("llmcli", llmcliArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "LLMCLI_MUTE_THINKING=1")
	var wg sync.WaitGroup
	if foldbin, err := exec.LookPath("fold"); err == nil {
		foldcmd := exec.Command(foldbin, "-s")
		if foldinput, err := foldcmd.StdinPipe(); err == nil {
			cmd.Stdout = foldinput
			foldcmd.Stdout = os.Stdout
			wg.Go(func() { foldcmd.Run() })
		}
	}
	var err error
	wg.Go(func() {
		defer cmd.Stdout.(io.Closer).Close() // fold will not exit until its standard input is closed
		err = cmd.Run()
	})
	wg.Wait()
	return err
}

func init() {
	flag.Usage = func() { fmt.Fprintln(flag.CommandLine.Output(), usage) }
}

const usage = "usage: tldr file|url [file|url]"
