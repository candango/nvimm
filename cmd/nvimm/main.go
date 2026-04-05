package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/candango/nvimm/internal/cli"
	"github.com/candango/nvimm/internal/config"
	"github.com/jessevdk/go-flags"
)

func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		return info.Main.Version
	}
	return "dev"
}

func main() {
	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-V" {
			fmt.Println(getVersion())
			os.Exit(0)
		}
	}

	var opts config.AppOptions

	parser := flags.NewParser(&opts, flags.Default)
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
		"List all Neovim versions currently installed and managed by nvimm on this machine.",
		&cli.ListCommand{})
	parser.AddCommand(
		"upgrade",
		"Upgrade Neovim to the latest stable version",
		"Check for the latest stable Neovim release and install it if not already present. Prompts before downloading and before setting as current. Use -y to skip all prompts.",
		&cli.UpgradeCommand{})

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && (flagsErr.Type == flags.ErrUnknownCommand || flagsErr.Type == flags.ErrUnknownFlag) {
			parser.WriteHelp(os.Stderr)
			os.Exit(1)
		}
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		parser.WriteHelp(os.Stderr)
		os.Exit(0)
	}
}
