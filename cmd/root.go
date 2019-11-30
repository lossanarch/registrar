package cmd

import (
	"github.com/lossanarch/registrar/pkg/registrar"
	"github.com/spf13/cobra"
)

var domain string

var rootCmd = &cobra.Command{
	Use:   "registrar <host>",
	Short: "Registers a network interface ip with route53",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		registrar.Register(args[0])

	},
}

func init() {

}

func Execute() {
	rootCmd.Execute()
}
