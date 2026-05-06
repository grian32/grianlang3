package main

import (
	"context"
	"embed"
	"grianlang3/cli"
	"os"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

//go:embed builtins/*.ll
var builtinFs embed.FS

func main() {
	cmd := &cobra.Command{
		Use:   "gl3",
		Short: "gl3 main command",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("You ran the root command. Try `gl3 build x.gl3`")
		},
	}

	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Compile gl3 files to an executable",
		RunE: func(cmd *cobra.Command, args []string) error {
			var buildOpts cli.BuildOpts
			cmd.Flags().BoolVar(&buildOpts.Dbg, "dbg", false, "Prints out the AST for all compiled files, along with the `clang` command used for compilation")
			cmd.Flags().BoolVar(&buildOpts.Dbg, "keepll", false, "Saves the .ll files produced by compilation")
			cmd.Flags().BoolVar(&buildOpts.Dbg, "noexecbuild", false, "Does not execute the `clang` build command")
			err := cli.RunBuildCmd(builtinFs, args, &buildOpts)
			if err != nil {
				if err := os.RemoveAll("./lltemp"); err != nil {
					return err
				}
			}
			return err
		},
	}

	cmd.AddCommand(buildCmd)

	if err := fang.Execute(
		context.Background(),
		cmd,
		fang.WithNotifySignal(os.Interrupt, os.Kill),
	); err != nil {
		os.Exit(1)
	}
}
