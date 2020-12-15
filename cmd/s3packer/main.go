package main

import (
	"fmt"
	"os"

	"github.com/koblas/s3packer/internal"
	"github.com/koblas/s3packer/version"
	"github.com/moby/term"
	"github.com/spf13/cobra"
)

func main() {
	cmd, err := buildCommand()

	_, stdout, stderr := term.StdStreams()

	if err != nil {
		fmt.Fprintf(stderr, "%s\n", err)

		os.Exit(1)
	}
	cmd.SetOut(stdout)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(stderr, "%s\n", err)

		os.Exit(1)
	}
}

func buildCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:           "s3packer DEST FILES...",
		Short:         "Pack a list of S3 files into a single zip",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return internal.ZipPack(args[0], args[1:])
		},
		DisableFlagsInUseLine: true,
		Version:               fmt.Sprintf("%s, build %s", version.Version, version.GitCommit),
	}

	return cmd, nil
}
