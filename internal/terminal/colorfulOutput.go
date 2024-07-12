package terminal

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"io"
	"regexp"
	"strings"
)

// ANSI color codes
var (
	yellow    = color.New(color.FgYellow).SprintFunc()
	blue      = color.New(color.FgBlue).SprintFunc()
	boldGreen = color.New(color.FgGreen, color.Bold).SprintFunc()
)

func ColorfulOutput(cmdOut io.Reader, printer io.Writer, promptStyle, promptSymbol string) error {
	reader := bufio.NewReader(cmdOut)

	buffer := make([]byte, 1024)

	for {
		read, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}

		if err == io.EOF {
			break
		}

		input := string(buffer[:read])

		// Regular expression to match the prompt pattern including ANSI escape sequences
		re := regexp.MustCompile(`(\x1b\)][^\a]*\a|\x1b\[\?2004h|\x1b\[2J)?([\w]+)@([\w.-]+):([^\s#$]+)[#$]`)

		matches := re.FindStringSubmatch(input)
		if matches != nil {
			prefix, user, path := matches[1], matches[2], matches[4]

			styledReplacement := prependPrefix(formatPrompt(promptStyle, user, path, promptSymbol), prefix)

			// Replace non styled text with the styled text in the original input
			styledInput := re.ReplaceAllString(input, styledReplacement)
			_, err := fmt.Fprint(printer, styledInput)
			if err != nil {
				return err
			}
		} else {
			_, err := fmt.Fprint(printer, input)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func formatPrompt(unparsedPrompt, user, workingDir, promptSymbol string) string {
	// Placeholder map
	placeholders := map[string]string{
		"\\u": blue(user),
		"\\p": boldGreen(promptSymbol),
		"\\w": yellow(workingDir),
	}

	// Replace placeholders in the format string
	parsedPrompt := unparsedPrompt
	for placeholder, value := range placeholders {
		parsedPrompt = strings.ReplaceAll(parsedPrompt, placeholder, value)
	}

	return parsedPrompt
}

func prependPrefix(text, prefix string) string {
	return prefix + text
}
