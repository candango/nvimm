package main

import (
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/candango/nvimim/internal/cli"
	"github.com/candango/nvimim/internal/config"
	"github.com/jessevdk/go-flags"
)

func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}

func main() {
	var opts config.AppOptions

	parser := flags.NewParser(&opts, flags.Default&^flags.PrintErrors)
	parser.Usage = "[Options] command"

	parser.CommandHandler = config.WithAppOptions(&opts, config.WithPathsResolved)

	parser.AddCommand(
		"current",
		"Display the active or installed Neovim version",
		"Show the version of Neovim currently in use or switch the active version to a specific installed build.",
		&cli.CurrentCommand{})
	parser.AddCommand(
		"install",
		"Install the latest or a specific Neovim version",
		"Download and install Neovim binaries directly from official releases. Supports 'latest', 'nightly', or specific version tags.",
		&cli.InstallCommand{})
	parser.AddCommand(
		"list",
		"List Neovim installed versions",
		"List all Neovim versions currently installed and managed by nvimim on this machine.",
		&cli.ListCommand{})
	parser.AddCommand(
		"upgrade",
		"Upgrade Neovim to the latest stable version",
		"Check for the latest stable Neovim release and install it if not already present. Prompts before downloading and before setting as current. Use -y to skip all prompts.",
		&cli.UpgradeCommand{})

	_, err := parser.Parse()
	if opts.Version {
		fmt.Println(getVersion())
		os.Exit(0)
	}
	if err != nil {
		if errors.Is(err, config.ErrVersionRequested) {
			fmt.Println(getVersion())
			os.Exit(0)
		}
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			parser.WriteHelp(os.Stdout)
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err)
		parser.WriteHelp(os.Stderr)
		os.Exit(1)
	}
}
