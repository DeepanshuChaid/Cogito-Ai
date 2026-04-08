package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use: "cogito",
		Short: "Cogito is a persistent memory layer for Ai",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Welcome to Cogito. Use --help for commands")
		},
	}

	var initCmd = &cobra.Command{
		Use: "init",
		Short: "Initialize a new Cogito Memo",
		Run: func(cmd *cobra.Command, args []string) {
			folderName := ".cogito"

			if _, err := os.Stat(folderName); os.IsNotExist(err) {
				err := os.Mkdir(folderName, 0755)
				if err != nil {
					fmt.Println("Error Creating a Directory", err)
					return
				}

				fmt.Println("📁 Created .cogito folder.")
			} else {
				fmt.Println("⚠️  .cogito folder already exists.")
			}

			filePath := folderName + "/memo.md"
			content := []byte("# Cogito Project Memo\nKeep track of your project state here.")

			err :=  os.WriteFile(filePath, content, 0644)
			if err != nil {
				fmt.Println("Error creating a memo file: ", err)
				return
			}
			fmt.Println("✅ Cogito initialized successfully!")
		},
	}

	rootCmd.AddCommand(initCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Error initing the library: ", err.Error())
		os.Exit(1)
	}
}
