package main

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/mail"
	"html/template"
	"io"
)

var templateFuncMap = template.FuncMap{
	"authProvider": func(authProviderStr string) string {
		return dao.AuthProvider(authProviderStr).HumanName()
	},
}

type NotificationProvider interface {
	Send(classifier string, notifications []dao.Notification) error
}

func NewLoggingNotificationProvider() NotificationProvider {
	templateData := MustAsset("logging-template")
	tmpl, err := template.New("report-email").Funcs(templateFuncMap).Parse(string(templateData))
	if err != nil {
		log.Fatal("Can not compile email template:\t", err)
	}
	return &loggingNotificationProvider{
		tmpl:&TemplateCache{
			preffix:"logging-template-",
			_default:tmpl,
			templates:make(map[string]*template.Template),
		},
	}
}

type loggingNotificationProvider struct {
	tmpl *TemplateCache
}

func (this *loggingNotificationProvider) Send(classifier string, notifications []dao.Notification) error {
	buf := bytes.Buffer{}
	this.tmpl.template(classifier).Execute(&buf, notifications)
	log.Infof("Messages for classifier %s:\n%s", classifier, buf.String())
	return nil
}

func NewEmailNotificationProvider(conf config.Notifications) NotificationProvider {
	templateData := MustAsset("email-template")
	tmpl, err := template.New("report-email").Funcs(templateFuncMap).Parse(string(templateData))
	if err != nil {
		log.Fatal("Can not compile email template:\t", err)
	}

	return &emailNotificationProvider{
		conf:conf,
		tmpl:&TemplateCache{
			preffix:"email-template-",
			_default:tmpl,
			templates:make(map[string]*template.Template),
		},
	}
}

type emailNotificationProvider struct {
	conf config.Notifications
	tmpl *TemplateCache
}

func (this *emailNotificationProvider) Send(classifier string, notifications []dao.Notification) error {
	recipients := make([]string, len(notifications))
	for i:=0;i<len(notifications);i++ {
		recipients[i] = notifications[i].Recipient.Recipient
	}
	return mail.SendMail(this.conf.MailSettings, recipients, this.conf.ReportingEmailSubject, func(w io.Writer) error {
		return this.tmpl.template(classifier).Execute(w, notifications)
	})
}

func NewVkNotificationProvider() NotificationProvider {
	return &vkNotificationProvider{}
}

type vkNotificationProvider struct {

}

func (this *vkNotificationProvider) Send(classifier string, notifications []dao.Notification) error {
	log.Error("VK provider is not implemented yet")
	return nil
}

type TemplateCache struct {
	preffix   string
	templates map[string]*template.Template
	_default  *template.Template
}

func (this *TemplateCache) template(classifier string) *template.Template {
	log.Debugf("Get template for classifier %s", classifier)
	tmpl, found := this.templates[classifier]
	if found && tmpl == nil {
		//means non-existing template - use default
		return this._default
	}
	if !found {
		templateName := this.preffix + classifier
		b, err := Asset(templateName)
		if err != nil {
			log.Errorf("Can not load template for cassifier %s: %v", classifier, err)
			this.templates[classifier] = nil
			return this._default
		}
		tmpl, err = template.New(templateName).Funcs(templateFuncMap).Parse(string(b))
		if err != nil {
			log.Errorf("Can not compile template for cassifier %s: %v", classifier, err)
			this.templates[classifier] = nil
			return this._default
		}
		this.templates[classifier] = tmpl
	}
	return tmpl
}
