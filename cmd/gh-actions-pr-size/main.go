package main

import (
	"context"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kkohtaka/gh-actions-pr-size/pkg/cmd"
	"github.com/spf13/cobra"
)

var (
	prSizeCmd *cobra.Command = cmd.PRSizeCmd
	exit      func(int)      = os.Exit
)

func main() {
	logger := zap.New(zap.ConsoleEncoder())
	ctx := log.IntoContext(context.Background(), zap.New())
	if err := prSizeCmd.ExecuteContext(ctx); err != nil {
		logger.Error(err, "Could not process the command.")
		exit(1)
	}
}
