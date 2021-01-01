package cmd

import (
	"crypto/tls"
	"fmt"
	"github.com/spf13/cobra"
	"github/ajanthan/smtp-go/pkg/api"
	"github/ajanthan/smtp-go/pkg/http"
	"github/ajanthan/smtp-go/pkg/smtp"
	"github/ajanthan/smtp-go/pkg/storage"
	"os"
)

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVarP(&ip, "address", "a", "127.0.0.1", "ip address of the smtp server")
	serverCmd.Flags().StringVarP(&pubKey, "certificate", "c", "", "public certificate of the server")
	serverCmd.Flags().StringVarP(&privateKey, "key", "k", "", "private key of the server")
	serverCmd.Flags().IntVarP(&smtpPort, "smtpPort", "m", 10587, "smtpPort of the smtp server")
	serverCmd.Flags().IntVarP(&httpPort, "httpPort", "u", 8085, "httpPort of the http server")
	serverCmd.Flags().BoolVarP(&secured, "secured", "s", true, "secure the SMTP communication")
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "server starts a smtp server on a network interface and smtpPort",
	Run: func(cmd *cobra.Command, args []string) {
		store, err := storage.NewStorage("mail.db")
		if err != nil {
			fmt.Println("Unable to initialize storage,", err.Error())
			os.Exit(1)
		}
		apiHandler := &api.MailAPI{Storage: *store}
		httpServer := &http.Server{
			Address:  ip,
			HTTPPort: httpPort,
		}
		fmt.Printf("starting a api server on %s:%d\n", ip, httpPort)
		go func() {
			err := httpServer.Start(apiHandler)
			if err != nil {
				fmt.Println("Unable to initialize http server,", err.Error())
				os.Exit(1)
			}
		}()
		smtpServer := &smtp.Server{
			Address:  ip,
			SMTPPort: smtpPort,
			Receiver: &storage.DBReceiver{
				Storage: store,
			},
		}
		if pubKey != "" && privateKey != "" {
			cert, err := tls.LoadX509KeyPair(pubKey, privateKey)
			if err != nil {
				fmt.Println("Unable load certificate,", err.Error())
				os.Exit(1)
			}
			smtpServer.TLSConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
		}
		if secured {
			smtpServer.AuthService = store
			smtpServer.Secure = secured
		}
		fmt.Printf("starting a smtp server on %s:%d\n", ip, smtpPort)
		smtpServer.Start()
	},
}

var ip string
var smtpPort int
var httpPort int
var pubKey string
var privateKey string
var secured bool
