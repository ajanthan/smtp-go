package http

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github/ajanthan/smtp-go/pkg/api"
	"net/http"
	"strconv"
)

type Server struct {
	Address  string
	HTTPPort int
}

func (s Server) Start(api *api.MailAPI) error {
	router := gin.Default()

	//CORS middleware configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost", "http://127.0.0.1"}
	router.Use(cors.New(corsConfig))

	router.LoadHTMLGlob("templates/*")

	router.GET("/mail", api.HandleGetAllMails)
	router.GET("/mail/:mailID/content", api.HandleGetMailByID)
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/app")
	})
	router.Static("/app", "./ui/app/xmail/build")

	err := router.Run(s.Address + ":" + strconv.Itoa(s.HTTPPort))
	if err != nil {
		return err
	}
	return nil
}
