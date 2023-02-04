// Copyright (C) 2023, MinIO, Inc.
//
// This code is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License, version 3,
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License, version 3,
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package main

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/minio/operator/pkg"

	"github.com/minio/cli"
	"github.com/minio/pkg/console"
	"github.com/minio/pkg/trie"
	"github.com/minio/pkg/words"
)

// Help template for Operator.
var operatorHelpTemplate = `NAME:
 {{.Name}} - {{.Usage}}

DESCRIPTION:
 {{.Description}}

USAGE:
 {{.HelpName}} {{if .VisibleFlags}}[FLAGS] {{end}}COMMAND{{if .VisibleFlags}}{{end}} [ARGS...]

COMMANDS:
 {{range .VisibleCommands}}{{join .Names ", "}}{{ "\t" }}{{.Usage}}
 {{end}}{{if .VisibleFlags}}
FLAGS:
 {{range .VisibleFlags}}{{.}}
 {{end}}{{end}}
VERSION:
 {{.Version}}
`

func newApp(name string) *cli.App {
	// Collection of console commands currently supported are.
	var commands []cli.Command

	// Collection of console commands currently supported in a trie tree.
	commandsTree := trie.NewTrie()

	// registerCommand registers a cli command.
	registerCommand := func(command cli.Command) {
		commands = append(commands, command)
		commandsTree.Insert(command.Name)
	}

	// register commands
	for _, cmd := range appCmds {
		registerCommand(cmd)
	}

	findClosestCommands := func(command string) []string {
		var closestCommands []string
		closestCommands = append(closestCommands, commandsTree.PrefixMatch(command)...)

		sort.Strings(closestCommands)
		// Suggest other close commands - allow missed, wrongly added and
		// even transposed characters
		for _, value := range commandsTree.Walk(commandsTree.Root()) {
			if sort.SearchStrings(closestCommands, value) < len(closestCommands) {
				continue
			}
			// 2 is arbitrary and represents the max
			// allowed number of typed errors
			if words.DamerauLevenshteinDistance(command, value) < 2 {
				closestCommands = append(closestCommands, value)
			}
		}

		return closestCommands
	}

	cli.HelpFlag = cli.BoolFlag{
		Name:  "help, h",
		Usage: "show help",
	}

	app := cli.NewApp()
	app.Name = name
	app.Version = pkg.Version + " - " + pkg.ShortCommitID
	app.Author = "MinIO, Inc."
	app.Usage = "MinIO Operator"
	app.Description = `MinIO Operator automates the orchestration of MinIO Tenants on Kubernetes.`
	app.Copyright = "(c) 2023 MinIO, Inc."
	app.Compiled, _ = time.Parse(time.RFC3339, pkg.ReleaseTime)
	app.Commands = commands
	app.HideHelpCommand = true // Hide `help, h` command, we already have `minio --help`.
	app.CustomAppHelpTemplate = operatorHelpTemplate
	app.CommandNotFound = func(ctx *cli.Context, command string) {
		console.Printf("‘%s’ is not a console sub-command. See ‘console --help’.\n", command)
		closestCommands := findClosestCommands(command)
		if len(closestCommands) > 0 {
			console.Println()
			console.Println("Did you mean one of these?")
			for _, cmd := range closestCommands {
				console.Printf("\t‘%s’\n", cmd)
			}
		}
		os.Exit(1)
	}

	return app
}

func main() {
	args := os.Args
	// Set the orchestrator app name.
	appName := filepath.Base(args[0])
	// Run the app - exit on error.
	if err := newApp(appName).Run(args); err != nil {
		os.Exit(1)
	}
}
