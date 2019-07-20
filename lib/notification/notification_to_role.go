package notification

import (
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/dao"
	"strings"
)

func YandexEmail(login string) string {
	if !strings.Contains(login, "@") {
		return login + "@yandex.ru"
	}
	return login
}

type NotificationHelper struct {
	FallbackEmailRecipient string
	UserDao                dao.UserDao
	NotificationDao        dao.NotificationDao
}

func (this *NotificationHelper) SendToRole(notificationTemplate dao.Notification, role dao.Role) error {
	notifications, err := this.NotificationToRole(notificationTemplate, role)
	if err != nil {
		log.Error("Can not create message for admins: ", err)
		notificationTemplate.Recipient = dao.NotificationRecipient{
			Provider:  dao.NOTIFICATION_PROVIDER_EMAIL,
			Recipient: this.FallbackEmailRecipient,
		}
		err2 := this.NotificationDao.Add(notificationTemplate)
		if err2 != nil {
			log.Error("Can not save notification: ", err2)
		}
		return err
	} else {
		err = this.NotificationDao.Add(notifications...)
		if err != nil {
			log.Error("Can not save notifications: ", err)
			return err
		}
	}
	return nil
}

func (this *NotificationHelper) NotificationToRole(notification dao.Notification, role dao.Role) ([]dao.Notification, error) {
	users, err := this.UserDao.ListByRole(role)
	if err != nil {
		return []dao.Notification{}, err
	}
	notifications := make([]dao.Notification, 0, len(users))
	for i := 0; i < len(users); i++ {
		if users[i].AuthProvider == dao.YANDEX {
			newNotification := notification
			newNotification.Recipient = dao.NotificationRecipient{
				Provider:  dao.NOTIFICATION_PROVIDER_EMAIL,
				Recipient: YandexEmail(users[i].Info.Login),
			}
			notifications = append(notifications, newNotification)
		}
	}
	return notifications, nil
}
