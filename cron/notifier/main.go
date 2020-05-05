package main

//go:generate go-bindata -pkg $GOPACKAGE -o templates.go -prefix templates/ ./templates

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	"time"
)

const MAX_MESSAGES int = 1000

func main() {
	log.Infof("Starting wwmap notification sender")
	configuration := config.Load("")
	configuration.ConfigureLogger()
	storage := dao.NewPostgresStorage(configuration.Db)

	notificator := Notificator{
		Conf:            configuration.Notifications,
		NotificationDao: dao.NewNotificationPostgresDao(storage),
		Providers:       make(map[dao.NotificationProvider]NotificationProvider),
	}
	notificator.DoSend()
}

type Notificator struct {
	Providers       map[dao.NotificationProvider]NotificationProvider
	Conf            config.Notifications
	NotificationDao dao.NotificationDao
}

func (this *Notificator) DoSend() {
	recipients, err := this.NotificationDao.ListUnreadRecipients(time.Now())
	if err != nil {
		log.Fatal("Can not select recipients:\t", err)
	}
	if len(recipients) == 0 {
		log.Info("No notifications to send found")
		return
	}

	for _, recipient := range recipients {
		log.Infof("Send notification to %v", recipient)

		notifications, err := this.NotificationDao.ListUnreadByRecipient(recipient, MAX_MESSAGES)
		if err != nil {
			log.Fatal("Can not select reports:\t", err)
		}

		provider, err := this.GetNotificationProvider(recipient.Provider)
		if err != nil {
			log.Errorf("Can not get notification provider for %s: %v", recipient.Provider, err)
			continue
		}

		err = provider.Send(recipient.Classifier, notifications)
		if err != nil {
			log.Errorf("Can not send notifications to %v: %v", recipient, err)
			continue
		}

		ids := make([]int64, len(notifications))
		for i := 0; i < len(notifications); i++ {
			ids[i] = notifications[i].Id
		}
		err = this.NotificationDao.MarkRead(ids)
		if err != nil {
			log.Fatal("Can not mark reports as read:\t", err)
		}
	}
}

func (this *Notificator) GetNotificationProvider(key dao.NotificationProvider) (NotificationProvider, error) {
	provider, found := this.Providers[key]
	if found {
		return provider, nil
	}
	provider, err := this.CreateNotificationProvider(key)
	if err == nil {
		this.Providers[key] = provider
	}
	return provider, err
}

func (this *Notificator) CreateNotificationProvider(key dao.NotificationProvider) (NotificationProvider, error) {
	switch key {
	case dao.NOTIFICATION_PROVIDER_EMAIL:
		return NewEmailNotificationProvider(this.Conf), nil
	case dao.NOTIFICATION_PROVIDER_VK:
		return NewVkNotificationProvider(), nil
	case dao.NOTIFICATION_PROVIDER_LOG:
		return NewLoggingNotificationProvider(), nil
	default:
		return nil, fmt.Errorf("Can not find notification provider for %s", key)
	}
}
