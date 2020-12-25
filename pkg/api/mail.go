package api

import (
	"github.com/gin-gonic/gin"
	"github/ajanthan/smtp-go/pkg/storage"
	"html/template"
	"net/http"
	"strconv"
)

type MailAPI struct {
	Storage storage.Storage
}

func (m MailAPI) HandleGetAllMails(context *gin.Context) {
	mails, err := m.Storage.GetAll()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"Message": err.Error()})
	}
	context.JSON(http.StatusOK, mails)
}
func (m MailAPI) HandleGetMailByID(context *gin.Context) {
	mailIDStr := context.Param("mailID")
	mailID, err := strconv.Atoi(mailIDStr)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"Message": err.Error()})
	}
	body, err := m.Storage.GetBodyByMailID(uint(mailID))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"Message": err.Error()})
	}
	contents := template.HTML(body.Data)
	context.HTML(http.StatusOK, "mail.tmpl", gin.H{"Data": contents})
}
