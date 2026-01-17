package main

import (
	"os"

	"github.com/candango/nvimm/internal/cli"
	"github.com/candango/nvimm/internal/config"
	"github.com/jessevdk/go-flags"
)

func main() {
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
