package main

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	tcs := []struct {
		name         string
		prSizeCmd    *cobra.Command
		wantExitCode int
	}{
		{
			name: "If the command succeeds, main() also succeeds.",
			prSizeCmd: &cobra.Command{
				RunE: func(cmd *cobra.Command, args []string) error {
					return nil
				},
			},
		},
		{
			name: "If the command returns with an error, main() exits with an error code.",
			prSizeCmd: &cobra.Command{
				RunE: func(cmd *cobra.Command, args []string) error {
					return fmt.Errorf("some reason")
				},
			},
			wantExitCode: 1,
		},
	}
	for _, tt := range tcs {
		t.Run(tt.name, func(t *testing.T) {
			origPRSizeCmd := prSizeCmd
			origExit := exit
			t.Cleanup(func() {
				prSizeCmd = origPRSizeCmd
				exit = origExit
			})
			prSizeCmd = tt.prSizeCmd
			exit = func(code int) {
				panic(code)
			}

			var gotExitCode int
			func() {
				defer func() {
					if msg := recover(); msg != nil {
						gotExitCode, _ = msg.(int)
					}
				}()
				main()
			}()
			if tt.wantExitCode > 0 {
				assert.Equal(t, gotExitCode, tt.wantExitCode)
			} else {
				assert.Zero(t, gotExitCode)
			}
		})
	}
}
