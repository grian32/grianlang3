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
	rootCmd := &cobra.Command{
		Use:   "gl3",
		Short: "gl3 main command",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("You ran the root command. Try `gl3 build x.gl3`")
		},
	}

	var buildOpts cli.BuildOpts
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Compile gl3 files to an executable",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cli.RunBuildCmd(builtinFs, args, &buildOpts)
			if err != nil {
				if err := os.RemoveAll("./lltemp"); err != nil {
					return err
				}
			}
			return err
		},
	}
	buildCmd.Flags().BoolVar(&buildOpts.Dbg, "dbg", false, "Prints out the AST for all compiled files, along with the `clang` command used for compilation")
	buildCmd.Flags().BoolVar(&buildOpts.Dbg, "keepll", false, "Saves the .ll files produced by compilation")
	buildCmd.Flags().BoolVar(&buildOpts.Dbg, "noexecbuild", false, "Does not execute the `clang` build command")

	var exDefOpts cli.ExDefOpts
	exDefCmd := &cobra.Command{
		Use:   "exdef",
		Short: "Extracts define statements from a C header file and outputs a .gl3 file defining them as constants",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cli.RunExDef(&exDefOpts)
		},
	}
	exDefCmd.Flags().StringVarP(&exDefOpts.InFile, "input", "i", "", "Path to the input header")
	exDefCmd.Flags().StringVarP(&exDefOpts.OutFile, "output", "o", "", "Path to the output .gl3 file")
	exDefCmd.MarkFlagRequired("input")
	exDefCmd.MarkFlagRequired("output")

	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(exDefCmd)

	if err := fang.Execute(
		context.Background(),
		rootCmd,
		fang.WithNotifySignal(os.Interrupt, os.Kill),
	); err != nil {
		os.Exit(1)
	}
}
