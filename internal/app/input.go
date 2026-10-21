package app

import "github.com/spf13/cobra"

func (a *App) readInput(cmd *cobra.Command, path string, max int64) ([]byte, error) {
	type result struct {
		b   []byte
		err error
	}
	done := make(chan result, 1)
	go func() { b, err := readInput(path, a.IO.In, max); done <- result{b, err} }()
	select {
	case <-cmd.Context().Done():
		return nil, cmd.Context().Err()
	case r := <-done:
		return r.b, r.err
	}
}
