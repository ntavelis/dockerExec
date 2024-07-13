package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/moby/term"
	"github.com/ntavelis/dockerExec/internal/containerd"
	"github.com/ntavelis/dockerExec/internal/handler"
	"github.com/ntavelis/dockerExec/internal/terminal"
)

const version = "0.0.7"
const usage = `Usage:

dockerExec <containerID> ~> Opens a bash session inside container and uses the default prompt style

dockerExec --shell=/bin/bash --user=root --promptStyle="\\u@\\w:\\p" --promptSymbol="$" <containerID> ~> Full usage with all supported flags, please read documentation for details

Flags:
  --shell        Specify the shell to use (default: /bin/bash)
  --user         Specify the user to run the shell as (default: current user)
  --promptStyle  Customize the prompt style (default: "👨\\u ~> 📂\\w\r\n\\p")
  --promptSymbol Customize the prompt symbol (default: ">")
  --help(-h)     Display this help message
  --version(-v)  Display version information`

func main() {
	// -------------------------------------------------------------------------
	// FLAGS
	shell := flag.String("shell", "/bin/bash", "Specify the shell to use")
	user := flag.String("user", "", "Specify the user to run the shell as")
	promptStyle := flag.String("promptStyle", "👨\\u ~> 📂\\w\r\n\\p", "Customize the prompt style")
	promptSymbol := flag.String("promptSymbol", ">", "Customize the prompt symbol")
	help := flag.Bool("help", false, "Display this help message")
	flag.BoolVar(help, "h", false, "Display this help message (long format)")
	ver := flag.Bool("v", false, "Print version information")
	flag.BoolVar(ver, "version", false, "Print version information (long format)")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usage)
	}
	flag.Parse()

	// Help is needed, print message
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	// Version is needed, print it
	if *ver {
		fmt.Fprintf(os.Stdout, "v%s\n", version)
		os.Exit(0)
	}

	// Check if the containerID argument is provided
	if len(flag.Args()) < 1 {
		fmt.Fprintln(os.Stderr, "Error: containerID is required")
		flag.Usage()
		os.Exit(1)
	}

	// The first argument after the flags is the containerID
	containerId := flag.Arg(0)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo,
	}))

	// -------------------------------------------------------------------------
	// LOGGER
	logger.Debug(fmt.Sprintf("Opening %s session inside container with id: %s", *shell, containerId))

	// -------------------------------------------------------------------------
	// RAW TERMINAL
	previousState, err := term.MakeRaw(os.Stdin.Fd())
	handleError(err, logger, previousState)
	defer term.RestoreTerminal(os.Stdin.Fd(), previousState)

	ctx := context.Background()

	if err := run(ctx, logger, containerId, previousState, unescapeString(*promptStyle), *promptSymbol, *shell, *user); err != nil {
		handleError(err, logger, previousState)
	}
}

func run(ctx context.Context, logger *slog.Logger, containerId string, state *term.State, promptStyle, promptSymbol, shell, user string) error {
	client, err := containerd.NewDefaultClient()
	if err != nil {
		return err
	}

	interactiveCommand, err := client.ExecInteractiveCommand(ctx, containerId, shell, user)
	if err != nil {
		return err
	}
	defer interactiveCommand.Writer.Close()

	go func() {
		err := handler.InputHandler(os.Stdin, interactiveCommand.Writer)
		if err != nil {
			handleError(err, logger, state)
		}
	}()

	outputDone := make(chan struct{}, 1)
	go func() {
		err := terminal.ColorfulOutput(interactiveCommand.Reader, os.Stdout, promptStyle, promptSymbol)
		if err != nil {
			handleError(err, logger, state)
		}
		close(outputDone)
	}()

	<-outputDone

	return nil
}

func handleError(err error, logger *slog.Logger, previousState *term.State) {
	if err != nil {
		// We need to exit return terminal to previous state
		// To properly print error messages
		term.RestoreTerminal(os.Stdin.Fd(), previousState)
		logger.Error(err.Error())
		os.Exit(1)
	}
}

// unescapeString Unescape promptStyle flag to properly parse specific special chars
func unescapeString(s string) string {
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\r`, "\r")
	s = strings.ReplaceAll(s, `\t`, "\t")
	return s
}
