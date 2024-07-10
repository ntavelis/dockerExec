package handler

import (
	"bufio"
	"io"
)

func InputHandler(stdIn io.Reader, writer io.Writer) error {
	stdinReader := bufio.NewReader(stdIn)
	for {
		readByte, err := stdinReader.ReadByte()
		if err != nil {
			return err
		}

		_, err = writer.Write([]byte{readByte})
		if err != nil {
			return err
		}
	}
}
