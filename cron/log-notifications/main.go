package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/backend/handler"
	"github.com/and-hom/wwmap/lib/config"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/notification"
	"time"
)

func main() {
	log.Infof("Starting wwmap notification sender")
	configuration := config.Load("")
	configuration.ChangeLogLevel()
	storage := dao.NewPostgresStorage(configuration.Db)

	logDao := dao.NewChangesLogPostgresDao(storage)
	userDao := dao.NewUserPostgresDao(storage)
	notificationDao := dao.NewNotificationPostgresDao(storage)
	notificationHelper := notification.NotificationHelper{
		FallbackEmailRecipient: configuration.Notifications.FallbackEmailRecipient,
		UserDao:                userDao,
		NotificationDao:        notificationDao,
	}

	now := time.Now()
	entries, err := logDao.ListAllTimeRange(now.Add(-24*time.Hour), now, 1000)
	if err != nil {
		log.Fatal("Can't list log entries: ", err)
	}
	if len(entries) == 0 {
		log.Info("No log entries for last 24 hours - exiting")
		return
	}

	eventsSummary := make(map[string]int)
	addedIdsByType := make(map[string][]int64)
	changedIdsByType := make(map[string][]int64)
	removedIdsByType := make(map[string][]int64)

	for _, entry := range entries {
		switch entry.Type {
		case dao.ENTRY_TYPE_CREATE:
			appendId(&addedIdsByType, entry.ObjectType, entry.ObjectId)
		case dao.ENTRY_TYPE_MODIFY:
			appendId(&changedIdsByType, entry.ObjectType, entry.ObjectId)
		case dao.ENTRY_TYPE_DELETE:
			appendId(&removedIdsByType, entry.ObjectType, entry.ObjectId)
		}
	}

	addEventSummaries("Добавлено", addedIdsByType, &eventsSummary)
	addEventSummaries("Изменено", changedIdsByType, &eventsSummary)
	addEventSummaries("Удалено", removedIdsByType, &eventsSummary)

	for msg, count := range eventsSummary {
		err = notificationHelper.SendToRole(dao.Notification{
			Object:     dao.IdTitle{0, msg},
			Comment:    fmt.Sprintf("%d", count),
			CreatedAt:  dao.JSONDate(now),
			Recipient:  dao.NotificationRecipient{}, // would be filled automatically for all admins
			Classifier: "editing-report",
			SendBefore: now, // send as soon as possible
		}, dao.ADMIN)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func addEventSummaries(actionStr string, addedIdsByType map[string][]int64, eventsSummary *map[string]int) {
	for t, ids := range addedIdsByType {
		if len(ids) == 0 {
			continue
		}
		objTitle, ok := objectTypeStr(t)
		if !ok {
			continue
		}
		msg := actionStr + " " + objTitle + ":"
		(*eventsSummary)[msg] = len(ids)
	}
}

func appendId(m *map[string][]int64, objectType string, id int64) {
	arr, found := (*m)[objectType]
	if !found {
		(*m)[objectType] = []int64{id}
	}
	for _, existing := range arr {
		if existing == id {
			return
		}
	}
	(*m)[objectType] = append(arr, id)
}

func objectActionStr(action dao.ChangesLogEntryType) string {
	switch action {
	case dao.ENTRY_TYPE_CREATE:
		return "добавлено"
	case dao.ENTRY_TYPE_MODIFY:
		return "отредактировано"
	case dao.ENTRY_TYPE_DELETE:
		return "удалено"
	}
	return string(action)
}

func objectTypeStr(objectType string) (string, bool) {
	switch objectType {
	case handler.RIVER_LOG_ENTRY_TYPE:
		return "Рек", true
	case handler.SPOT_LOG_ENTRY_TYPE:
		return "Порогов", true
	case handler.IMAGE_LOG_ENTRY_TYPE:
		return "Изображений", true
	}
	return "", false
}
