package main

import (
	"context"
	"dockerExec/internal/containerd"
	"dockerExec/internal/handler"
	"dockerExec/internal/terminal"
	"flag"
	"fmt"
	"github.com/moby/term"
	"log/slog"
	"os"
)

func main() {
	// -------------------------------------------------------------------------
	// FLAGS
	shell := flag.String("shell", "/bin/bash", "The shell to use")
	user := flag.String("user", "", "The user to login with")
	promptStyle := flag.String("promptStyle", "👨\\u ~> 📂\\w\r\n\\p", "The prompt style to use")
	promptSymbol := flag.String("promptSymbol", ">", "The prompt symbol to use")
	flag.Parse()

	// Check if the containerID argument is provided
	if len(flag.Args()) < 1 {
		fmt.Println("Usage: program --shell=/bin/bash --user=root --promptStyle=\"\\u@\\w:\\p\" --promptSymbol=\"\\$\" <containerID>")
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

	if err := run(ctx, logger, containerId, previousState, *promptStyle, *promptSymbol, *shell, *user); err != nil {
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
