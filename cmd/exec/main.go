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
		fmt.Println("Usage: program --shell=<shell_type> --user=<user> <containerID>")
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
		err := terminal.ColorfulOutput(interactiveCommand.Reader, os.Stdout)
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

// colorfulOutput reads from cmdOut using a buffer and outputs colorful terminal-like text.
//func colorfulOutput(cmdOut io.Reader, promptStyle, promptSymbol string) error {
//	reader := bufio.NewReader(cmdOut)
//
//	buffer := make([]byte, 1024)
//
//	for {
//		read, err := reader.Read(buffer)
//		if err != nil && err != io.EOF {
//			return err
//		}
//
//		if err == io.EOF {
//			break
//		}
//
//		input := string(buffer[:read])
//
//		// Regular expression to match the pattern
//		re := regexp.MustCompile(`(?:\x1b\[\?2004h|\x1b\[H\x1b\[2J)?([\w]+)@([\w.-]+):([^#$]+)[#$]`)
//
//		// Find the matchesunparsedPrompt
//		matches := re.FindStringSubmatch(input)
//		if matches != nil {
//			// Extract user, host, and path
//			user, path := matches[1], matches[3]
//
//			// Construct the styled output
//			// Construct the styled replacement
//			styledReplacement := prependEscapeChar(formatPrompt(promptStyle, user, path, promptSymbol))
//
//			// Replace the unstyled text with the styled text in the original input
//			styledInput := re.ReplaceAllString(input, styledReplacement)
//			_, err = fmt.Fprint(os.Stdout, styledInput)
//			if err != nil {
//				return err
//			}
//		} else {
//			_, err = fmt.Fprint(os.Stdout, input)
//			if err != nil {
//				return err
//			}
//		}
//	}
//
//	return nil
//}

//func formatPrompt(unparsedPrompt, user, workingDir, promptSymbol string) string {
//	// Placeholder map
//	placeholders := map[string]string{
//		"\\u": blue(user),
//		"\\p": boldGreen(promptSymbol),
//		"\\w": yellow(workingDir),
//	}
//
//	// Replace placeholders in the format string
//	parsedPrompt := unparsedPrompt
//	for placeholder, value := range placeholders {
//		parsedPrompt = strings.ReplaceAll(parsedPrompt, placeholder, value)
//	}
//
//	return parsedPrompt
//}
//
//func prependEscapeChar(text string) string {
//	// add escape char for terminal
//	return "\u001B[?2004h" + text
//}
