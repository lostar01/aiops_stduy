/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// askCmd represents the ask command
var askCmd = &cobra.Command{
	Use:   "ask",
	Short: "command for ask",
	Long: `For example:
	ask chatgpt
	`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ask called")
	},
}

func init() {
	rootCmd.AddCommand(askCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// askCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// askCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
