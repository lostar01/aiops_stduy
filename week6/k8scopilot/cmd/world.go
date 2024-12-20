/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// worldCmd represents the world command
var worldCmd = &cobra.Command{
	Use:        "world",
	Short:      "an demo for sub-command",
	Long:       `This is example for cobra sub command`,
	Deprecated: "This command is deprecated, please use hello instead",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("kubeconfig:", kubeconfig)
		fmt.Println("namespapce:", namespace)
		fmt.Println("source:", source)
	},
}

var source string

func init() {
	helloCmd.AddCommand(worldCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// worldCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// worldCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	worldCmd.Flags().StringVarP(&source, "source", "s", "world", "The source of the message")
}
