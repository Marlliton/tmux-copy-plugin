package cmd

import (
	"github.com/Marlliton/tmux-copy-plugin/internal/app"
	"github.com/Marlliton/tmux-copy-plugin/internal/config"
	"github.com/spf13/cobra"
)

var cfg config.Config

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Starts the plugin using the default behavior.",
	Long: `Watches Tmux for completed copy operations and forwards the captured text
to the system clipboard, ensuring multi-line copies work even inside Neovim.
You can define a notification type:
	preview, msg, system or none`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return cfg.Validate()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return app.Run(cfg)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP((*string)(&cfg.NotificationStyle), "notify", "n", "preview", "render a copy preview.")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
