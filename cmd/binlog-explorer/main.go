package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"db-log-explorer/internal/explorer"
	"db-log-explorer/internal/ui"
)

const version = "0.2.0"

func main() {
	help := flag.Bool("help", false, "show help")
	helpShort := flag.Bool("h", false, "show help")
	showVersion := flag.Bool("version", false, "show version")
	showVersionShort := flag.Bool("v", false, "show version")
	fromFlag := flag.String("from", "", "investigation scope start (inclusive); requires --to")
	toFlag := flag.String("to", "", "investigation scope end (inclusive); requires --from")
	flag.Parse()

	if *help || *helpShort {
		printUsage()
		os.Exit(0)
	}
	if *showVersion || *showVersionShort {
		fmt.Println(version)
		os.Exit(0)
	}

	if !term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Fprintln(os.Stderr, "error: terminal (TTY) required")
		os.Exit(1)
	}

	fromSet := strings.TrimSpace(*fromFlag) != ""
	toSet := strings.TrimSpace(*toFlag) != ""
	if fromSet != toSet {
		fmt.Fprintln(os.Stderr, "error: --from and --to must both be provided or both omitted")
		os.Exit(1)
	}

	session := explorer.NewSession()
	if fromSet {
		scope, err := explorer.ParseLaunchScope(*fromFlag, *toFlag)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		session.SetLaunchScope(scope)
	}

	paths := flag.Args()
	exitCode := 0
	var openOK, openFail int
	for _, p := range paths {
		if err := session.OpenSource(p); err != nil {
			openFail++
			session.AddLaunchWarning(fmt.Sprintf("%s: %v", p, err))
			continue
		}
		openOK++
	}

	if len(paths) > 0 && openOK == 0 {
		fmt.Fprintln(os.Stderr, "error: failed to open all binlog files")
		for _, w := range session.LaunchWarnings {
			fmt.Fprintln(os.Stderr, w)
		}
		os.Exit(1)
	}
	if openFail > 0 && openOK > 0 {
		exitCode = 2
	}

	model := ui.NewModel(session, exitCode)
	p := tea.NewProgram(&model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(model.ExitCode())
}

func printUsage() {
	fmt.Println("binlog-explorer [flags] [binlog-file ...]")
	fmt.Println()
	fmt.Println("Flags:")
	flag.PrintDefaults()
}
