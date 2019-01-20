package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/tanema/wrp/src/config"
)

const version = "wrp 0.0.1"

var cfg *config.Config

var WrpCmd = &cobra.Command{
	Use:   "wrp",
	Short: "wrp is a git based package dependency manager",
	Long: `wrp is a git based package dependency manager
It is probably not suitable for many people except me.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.CalledAs() == "init" {
			cfg = config.Default
			return nil
		}
		var err error
		cfg, err = config.Parse()
		if err != nil && err == config.ErrNotFound && cmd.CalledAs() == "add" {
			cfg = config.Default
			return nil
		}
		return err
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return cfg.FetchAllDependencies(false)
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		return cfg.Save()
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of wrp",
	Long:  `All software has versions. This is wrp's version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("wrp %s %s/%s", version, runtime.GOOS, runtime.GOARCH)
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init config",
	Long:  `init config`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install dependencies",
	Long:  `install dependencies`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cfg.FetchAllDependencies(false)
	},
}

var updateCmd = &cobra.Command{
	Use:   "update [REPO]",
	Short: "update a dependency in project",
	Long:  "update a dependency in project",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cfg.FetchAllDependencies(true)
		}
		return cfg.Update(args[0])
	},
}

var removeCmd = &cobra.Command{
	Use:     "rm [REPO]",
	Aliases: []string{"remove"},
	Short:   "rm a dependency from project",
	Long:    "rm a dependency from project",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cfg.Remove(args[0])
	},
}

var addCmd = &cobra.Command{
	Use:   "add [REPO] [PICKS]",
	Short: "add a new dependency to project",
	Long:  "add a new dependency to project",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cfg.Add(args[0], args[1:])
	},
}

func init() {
	WrpCmd.AddCommand(initCmd, addCmd, removeCmd, updateCmd, versionCmd)
}
