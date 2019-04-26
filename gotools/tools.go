// +build tools

package tools

import (
	// golint is a linter for Go source code.
	_ "golang.org/x/lint/golint"

	// counterfeiter is a tool for generatinng mocks from interfaces.
	_ "github.com/maxbrunsfeld/counterfeiter/v6"
)
