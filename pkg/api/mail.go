package api

import (
	"github.com/gin-gonic/gin"
	"github/ajanthan/smtp-go/pkg/storage"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type MailAPI struct {
	Storage storage.SQLiteStorage
}

func (m MailAPI) HandleGetAllMails(context *gin.Context) {
	mails, err := m.Storage.GetAll()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"Message": err.Error()})
		return
	}
	context.JSON(http.StatusOK, mails)
}
func (m MailAPI) HandleGetMailByID(context *gin.Context) {
	mailIDStr := context.Param("mailID")
	mailID, err := strconv.Atoi(mailIDStr)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"Message": err.Error()})
		return
	}
	body, err := m.Storage.GetBodyByMailID(uint(mailID))
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"Message": err.Error()})
		return
	}
		if strings.HasPrefix(body.ContentType, "text/html;") {
			contents := template.HTML(body.Data)
			context.HTML(http.StatusOK, "mail.tmpl", gin.H{"Data": contents})
		} else {
			context.String(http.StatusOK, string(body.Data))
		}
}
