package main

import (
	log "github.com/Sirupsen/logrus"
	. "github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/config"
	"html/template"
	"io"
	"github.com/and-hom/wwmap/lib/mail"
)

const MAX_MESSAGES int = 1000

func main() {
	log.Infof("Starting wwmap notification sender")
	configuration := config.Load("")
	storage := NewPostgresStorage(configuration.DbConnString)
	reportStorage := NewNotificationPostgresDao(storage)

	reports, err := reportStorage.ListUnread(MAX_MESSAGES)
	if err != nil {
		log.Fatal("Can not select reports:\t", err)
	}
	if len(reports) == 0 {
		log.Info("No reports to send found")
		return
	}

	templateData, err := emailTemplateBytes()
	if err != nil {
		log.Fatal("Can not load email template:\t", err)
	}

	tmpl, err := template.New("report-email").Parse(string(templateData))
	if err != nil {
		log.Fatal("Can not compile email template:\t", err)
	}

	err = mail.SendMail(configuration.Notifications.EmailSender, configuration.Notifications.EmailRecipients, configuration.Notifications.ReportingEmailSubject, func(w io.Writer) error {
		return tmpl.Execute(w, reports)
	})
	if err != nil {
		log.Fatal("Can not send emails:\t", err)
	}

	ids := make([]int64, len(reports))
	for i := 0; i < len(reports); i++ {
		ids[i] = reports[i].Id
	}
	err = reportStorage.MarkRead(ids)
	if err != nil {
		log.Fatal("Can not mark reports as read:\t", err)
	}

}
