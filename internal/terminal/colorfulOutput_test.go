package terminal

import (
	"bytes"
	"io"
	"testing"
)

func TestColorfulOutput(t *testing.T) {
	type args struct {
		cmdOut       io.Reader
		promptStyle  string
		promptSymbol string
	}
	tests := []struct {
		name        string
		args        args
		wantPrinter string
		wantErr     bool
	}{
		{
			"ItWillHandleBashPrompt",
			args{
				bytes.NewBufferString("\u001B[?2004hroot@4ac0a1a3eb6c:/# "),
				"👨 \\u ~> 📂\\w\r\n\\p",
				">",
			},
			"\u001B[?2004h👨 root ~> 📂/\r\n> ",
			false,
		},
		{
			"ItWillHandleBashPrompt",
			args{
				bytes.NewBufferString("]0;root@e80dbcfcaa55: /srv/app\aroot@e80dbcfcaa55:/srv/app#"),
				"👨 \\u ~> 📂\\w\r\n\\p",
				">",
			},
			"]0;root@e80dbcfcaa55: /srv/app\a👨 root ~> 📂/srv/app\r\n>",
			false,
		},
		{
			"ItWillHandleBashPromptInsideBigText",
			args{
				bytes.NewBufferString("bin   docker-entrypoint.d   home   media  proc\tsbin  tmp\nboot  docker-entrypoint.sh  lib    mnt\t  root\tsrv   usr\ndev   etc\t\t    lib64  opt\t  run\tsys   var\n\u001B[?2004hroot@f29ec6f0d5a5:/# "),
				"👨 \\u ~> 📂\\w\r\n\\p",
				">",
			},
			"bin   docker-entrypoint.d   home   media  proc\tsbin  tmp\nboot  docker-entrypoint.sh  lib    mnt\t  root\tsrv   usr\ndev   etc\t\t    lib64  opt\t  run\tsys   var\n\u001B[?2004h👨 root ~> 📂/\r\n> ",
			false,
		},
		{
			"IfWeAreUnableToParseItWePrintItAsIsBinShPrompt",
			args{
				bytes.NewBufferString("#"),
				"👨 \\u ~> 📂\\w\r\n\\p",
				">",
			},
			"#",
			false,
		},
		{
			"IfWeAreUnableToParseItWePrintItAsIsEnterKey",
			args{
				bytes.NewBufferString("\n# "),
				"👨 \\u ~> 📂\\w\r\n\\p",
				">",
			},
			"\n# ",
			false,
		},
		{
			"ItCanHaveCustomPromptStyleAndPromptSymbol",
			args{
				bytes.NewBufferString("\u001B[?2004hroot@4ac0a1a3eb6c:/# "),
				"\\u@\\w:\\p",
				"$",
			},
			"\u001B[?2004hroot@/:$ ",
			false,
		},
		{
			"ItWillIgnoreRandomEscapeChars",
			args{
				bytes.NewBufferString("\u001B[?2004hroot@4ac0a1a3eb6c:/# "),
				"\\u@\\w:\\p\\r\\t",
				"$",
			},
			"\u001B[?2004hroot@/:$\\r\\t ",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			printer := &bytes.Buffer{}
			err := ColorfulOutput(tt.args.cmdOut, printer, tt.args.promptStyle, tt.args.promptSymbol)
			if (err != nil) != tt.wantErr {
				t.Errorf("ColorfulOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPrinter := printer.String(); gotPrinter != tt.wantPrinter {
				t.Errorf("ColorfulOutput() gotPrinter = %v, want %v", gotPrinter, tt.wantPrinter)
			}
		})
	}
}
