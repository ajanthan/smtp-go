package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "smtp",
	Short: "smtp is command line to send or receive email via smtp protocol",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
