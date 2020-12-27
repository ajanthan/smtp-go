package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github/ajanthan/smtp-go/pkg/smtp"
	"github/ajanthan/smtp-go/pkg/storage"
	"os"
	"strconv"
	"strings"
)

var serverAddress string
var sender string
var subject string
var recipient string
var message string

func init() {
	rootCmd.AddCommand(client)
	client.AddCommand(getMails)
	client.AddCommand(sendMail)
	sendMail.Flags().StringVarP(&serverAddress, "address", "a", "127.0.0.1:10587", "address:smtpPort of the smtp server")
	sendMail.Flags().StringVarP(&subject, "subject", "s", "", "subject of the email")
	sendMail.Flags().StringVarP(&sender, "sender", "t", "", "email address of sender")
	sendMail.Flags().StringVarP(&recipient, "recipient", "r", "", "email address of recipient")
	sendMail.Flags().StringVarP(&message, "message", "m", "", "message")
}

var client = &cobra.Command{
	Use:   "client",
	Short: "client to interact with smtp server",
}
var sendMail = &cobra.Command{
	Use:   "send",
	Short: "client sends to email",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sending an email...")
		err := smtp.SendEmail(serverAddress, sender, recipient, subject, message)
		if err != nil {
			fmt.Printf("Unable to send the email: %s\n", err.Error())
		}
	},
}

var getMails = &cobra.Command{
	Use:   "get",
	Short: "get all email",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := storage.NewStorage("mail.db")
		if err != nil {
			fmt.Println("Unable to initialize storage,", err.Error())
			os.Exit(1)
		}
		if len(args) > 0 {
			fmt.Println("getting an email content...")
			id, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Println("invalid mail id,", args[0])
				os.Exit(1)
			}
			body, err := store.GetBodyByMailID(uint(id))
			if err != nil {
				fmt.Println("unable to get the mail content fot id=,", args[0])
				os.Exit(1)
			}
			fmt.Printf("Message:%s", string(body.Data))
		} else {
			fmt.Println("getting all emails...")
			emails, err := store.GetAll()
			if err != nil {
				fmt.Printf("Unable to get all the emails: %s\n", err.Error())
			}

			printDivider()
			fmt.Printf("|%-3s|", "ID")
			fmt.Printf("%-20s|", "TO")
			fmt.Printf("%-30s|", "FROM")
			fmt.Printf("%-40s|", "SUBJECT")
			fmt.Printf("%-40s|\n", "URL")
			for _, email := range emails {
				printDivider()
				fmt.Printf("|%-3d|", email.ID)
				if len(email.To[0]) > 20 {
					fmt.Printf("%-20s..|", email.To[0][0:18])
				} else {
					fmt.Printf("%-20s|", email.To[0])
				}
				if len(email.From) > 30 {
					fmt.Printf("%-30s..|", email.From[0:28])
				} else {
					fmt.Printf("%-30s|", email.From)
				}
				if len(email.Subject) > 40 {
					fmt.Printf("%-40s..|", email.Subject[0:38])
				} else {
					fmt.Printf("%-40s|", email.Subject)
				}
				fmt.Printf("%-40v|\n", " http://localhost:8085/mail/"+strconv.Itoa(int(email.ID))+"/content")
			}
			printDivider()
		}
	},
}

func printDivider() {
	fmt.Printf("+%-3s+", strings.Repeat("-", 3))
	fmt.Printf("%-20s+", strings.Repeat("-", 20))
	fmt.Printf("%-30s+", strings.Repeat("-", 29))
	fmt.Printf("%-40s+", strings.Repeat("-", 40))
	fmt.Printf("%-40s+\n", strings.Repeat("-", 40))
}
