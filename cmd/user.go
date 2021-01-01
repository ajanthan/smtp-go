package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github/ajanthan/smtp-go/pkg/storage"
	"os"
)

func init() {
	rootCmd.AddCommand(user)
	user.AddCommand(addUser)
	addUser.Flags().StringVarP(&username, "username", "u", "", "username for the user")
	addUser.Flags().StringVarP(&password, "password", "p", "", "password for the user")
}

var user = &cobra.Command{
	Use:   "user",
	Short: "admin users",
}
var addUser = &cobra.Command{
	Use:   "add",
	Short: "add a user",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := storage.NewStorage("mail.db")
		if err != nil {
			fmt.Println("Unable to initialize storage,", err.Error())
			os.Exit(1)
		}
		hmacSecret, err := store.AddUser(username, []byte(password))
		if err != nil {
			fmt.Println("Unable to add user,", err.Error())
			os.Exit(1)
		}
		fmt.Println("added the user successfully")
		fmt.Printf("HMAC secret: %s", hmacSecret)
	},
}

var username, password string
