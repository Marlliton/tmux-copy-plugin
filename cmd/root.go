package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tmux-copy-plugin",
	Short: "A CLI tool that fixes Tmux copy issues inside Neovim terminals.",
	Long: `A small utility that monitors Tmux copy events and reliably sends copied text
to the system clipboard, avoiding conflicts that occur when Tmux runs inside
Neovim's integrated terminal.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
