package main

import (
	"debug/buildinfo"
	"fmt"
	"log"
	"os"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/spf13/cobra"
	"github.com/wreulicke/waffle"
)

func mainInternal() error {
	//nolint:wrapcheck
	return NewApp().Execute()
}

func main() {
	if err := mainInternal(); err != nil {
		log.Fatal(err)
	}
}

func NewApp() *cobra.Command {
	var template string
	var output string
	c := cobra.Command{
		Use:   "waffle",
		Short: "template generator",
		RunE: func(cmd *cobra.Command, args []string) error {
			t := waffle.OpenTemplate(template)
			fs := osfs.New(output)
			return t.Generate(fs)
		},
	}
	c.Flags().StringVarP(&template, "template", "t", "", "template directory")
	c.Flags().StringVarP(&output, "output", "o", "", "output directory")
	_ = c.MarkFlagRequired("template")
	_ = c.MarkFlagRequired("output")

	c.AddCommand(
		NewVersionCommand(),
	)
	return &c
}

func NewVersionCommand() *cobra.Command {
	var detail bool
	c := &cobra.Command{
		Use:   "version",
		Short: "show version",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			info, err := buildinfo.ReadFile(os.Args[0])
			if err != nil {
				return fmt.Errorf("Cannot read buildinfo: %w", err)
			}

			fmt.Fprintf(w, "go version: %s", info.GoVersion)
			fmt.Fprintf(w, "module version: %s", info.Main.Version)
			if detail {
				fmt.Fprintln(w)
				fmt.Fprintln(w, info)
			}
			return nil
		},
	}
	c.Flags().BoolVarP(&detail, "detail", "d", false, "show details")
	return c
}
