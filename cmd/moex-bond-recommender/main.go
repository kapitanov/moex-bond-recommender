package main

import (
	"io"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCommand = &cobra.Command{
	Use:              "moex-bond-recommender",
	TraverseChildren: true,
	SilenceUsage:     true,
}

var quietMode bool
var appLogger *log.Logger

func init() {
	rootCommand.PersistentFlags().BoolVarP(&quietMode, "quiet", "q", false, "Suppress logging")

	rootCommand.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		log.SetFlags(log.LstdFlags | log.Lmsgprefix)
		if quietMode {
			log.SetOutput(io.Discard)
		} else {
			log.SetOutput(os.Stderr)
		}

		appLogger = log.New(log.Writer(), "app:  ", log.Flags())
	}
}

func main() {
	err := rootCommand.Execute()
	if err != nil {
		os.Exit(-1)
	}
}
