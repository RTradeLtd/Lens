package server

import (
	"testing"

	"go.uber.org/zap"

	"github.com/RTradeLtd/Lens/logs"
)

func Test_options(t *testing.T) {
	var l, _ = logs.NewLogger("", false)
	type args struct {
		certpath string
		keypath  string
		token    string
		logger   *zap.SugaredLogger
	}
	tests := []struct {
		name        string
		args        args
		wantOptions int
		wantErr     bool
	}{
		{"token too short",
			args{"", "", "", l}, 0, true},
		{"no logger provided",
			args{"", "", "asdfasdf", nil}, 0, true},
		{"invalid tls",
			args{"../README.md", "", "asdfasdf", l}, 0, true},
		{"ok: no tls",
			args{"", "", "asdfasdf", l}, 2, false},
		// disabled for now
		/*
			{"ok: with tls",
				args{"../test/certs/crt", "../test/certs/key", "asdfasdf", l}, 3, false},
		*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := options(tt.args.certpath, tt.args.keypath, tt.args.token, tt.args.logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("options() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.wantOptions {
				t.Errorf("expected %d options, got %d", tt.wantOptions, len(got))
			}
		})
	}
}
