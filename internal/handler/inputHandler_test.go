package handler

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestInputHandler(t *testing.T) {
	type args struct {
		stdIn io.Reader
	}
	tests := []struct {
		name       string
		args       args
		wantWriter string
		wantErr    bool
	}{
		{name: "TestItWillWriteToTheWriter", args: args{stdIn: strings.NewReader("hello world")}, wantWriter: "hello world", wantErr: false},
		{name: "TestItWillWriteToTheWriterSmallText", args: args{stdIn: strings.NewReader("l")}, wantWriter: "l", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := &bytes.Buffer{}
			err := InputHandler(tt.args.stdIn, writer)
			if (err != nil && err.Error() != "EOF") != tt.wantErr {
				t.Errorf("InputHandler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotWriter := writer.String(); gotWriter != tt.wantWriter {
				t.Errorf("InputHandler() gotWriter = %v, want %v", gotWriter, tt.wantWriter)
			}
		})
	}
}
