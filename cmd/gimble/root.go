package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:               "gimble",
		Short:             "Run coding agents from ordinary Go workflows",
		SilenceErrors:     true,
		SilenceUsage:      true,
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}
	root.AddCommand(newRunCommand(), newLsCommand(), newRunPromptCommand())
	return root
}

func absoluteDirectory(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve workdir: %w", err)
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return "", fmt.Errorf("inspect workdir: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("workdir %q is not a directory", absolute)
	}
	return absolute, nil
}
