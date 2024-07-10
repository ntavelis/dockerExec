package terminal

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"io"
	"regexp"
)

// ANSI color codes
var (
	yellow    = color.New(color.FgYellow).SprintFunc()
	blue      = color.New(color.FgBlue).SprintFunc()
	boldGreen = color.New(color.FgGreen, color.Bold).SprintFunc()
)

func ColorfulOutput(cmdOut io.Reader, printer io.Writer) error {
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
		re := regexp.MustCompile(`(\x1b\)][^\a]*\a|\x1b\[\?2004h)?([\w]+)@([\w.-]+):([^\s#$]+)[#$]`)

		matches := re.FindStringSubmatch(input)
		if matches != nil {
			prefix, user, path := matches[1], matches[2], matches[4]

			styledReplacement := fmt.Sprintf("%s👨 %s ~> 📂%s\r\n%s", prefix, blue(user), yellow(path), boldGreen(">"))

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
