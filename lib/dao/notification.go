package dao

import (
	"database/sql"
	"github.com/and-hom/wwmap/lib/dao/queries"
	"github.com/lib/pq"
	"time"
)

func NewNotificationPostgresDao(postgresStorage PostgresStorage) NotificationDao {
	return notificationStorage{
		PostgresStorage:       postgresStorage,
		insertQuery:           queries.SqlQuery("notification", "insert"),
		listUnreadQuery:       queries.SqlQuery("notification", "list-unread"),
		unreadRecipientsQuery: queries.SqlQuery("notification", "unread-provider-recipient-classifier"),
		markReadQuery:         queries.SqlQuery("notification", "mark-read"),
	}
}

type notificationStorage struct {
	PostgresStorage
	insertQuery           string
	listUnreadQuery       string
	unreadRecipientsQuery string
	markReadQuery         string
}

func (this notificationStorage) Add(notifications ...Notification) error {
	params := make([]interface{}, len(notifications))
	for i := 0; i < len(notifications); i++ {
		params[i] = notifications[i]
	}

	_, err := this.updateReturningId(this.insertQuery, func(entity interface{}) ([]interface{}, error) {
		notification := entity.(Notification)
		return []interface{}{notification.Title,
			notification.Object.Id, notification.Object.Title, notification.Comment,
			notification.Recipient.Provider, notification.Recipient.Recipient, notification.Classifier, notification.SendBefore}, nil
	}, true, params...)
	return err
}

func (this notificationStorage) ListUnreadRecipients(nowTime time.Time) ([]NotificationRecipientWithClassifier, error) {
	rwcs, err := this.doFindList(this.unreadRecipientsQuery,
		func(rows *sql.Rows) (NotificationRecipientWithClassifier, error) {
			notification := NotificationRecipientWithClassifier{}
			err := rows.Scan(&notification.Provider, &notification.Recipient, &notification.Classifier)
			return notification, err
		}, nowTime)
	if err != nil {
		return []NotificationRecipientWithClassifier{}, err
	}
	return rwcs.([]NotificationRecipientWithClassifier), err
}

func (this notificationStorage) ListUnreadByRecipient(rc NotificationRecipientWithClassifier, limit int) ([]Notification, error) {
	notifications, err := this.doFindList(this.listUnreadQuery,
		func(rows *sql.Rows) (Notification, error) {
			notification := Notification{}
			err := rows.Scan(&notification.Id, &notification.Title, &notification.Object.Id, &notification.Object.Title,
				&notification.Comment, &notification.CreatedAt, &notification.Recipient.Provider, &notification.Recipient.Recipient,
				&notification.Classifier, &notification.SendBefore)
			return notification, err
		}, limit, rc.Provider, rc.Recipient, rc.Classifier)
	if err != nil {
		return []Notification{}, err
	}
	return notifications.([]Notification), err
}

func (this notificationStorage) MarkRead(ids []int64) error {
	return this.performUpdates(this.markReadQuery,
		func(ids interface{}) ([]interface{}, error) {
			return []interface{}{ids}, nil
		}, pq.Array(ids))
}
