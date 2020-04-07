package dao

import (
	"database/sql"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"reflect"
)

type Storage interface {
	// Call payload function within transaction if supported by storage. Simply call payload function if not supported.
	WithinTx(payload func(tx interface{}) error) error
}

func (this *PostgresStorage) Begin() (PgTxHolder, error) {
	tx, err := this.db.Begin()
	if err != nil {
		return PgTxHolder{}, err
	}
	return PgTxHolder{PostgresStorage: *this, tx: tx}, nil
}

func (this PostgresStorage) WithinTx(payload func(tx interface{}) error) error {
	tx, err := this.Begin()
	if err != nil {
		return err
	}
	defer tx.Close()

	err = payload(tx)

	if err != nil {
		return err
	}
	return tx.Commit()
}

type PgTxHolder struct {
	PostgresStorage
	tx       *sql.Tx
	commited bool
}

func (this *PgTxHolder) Close() error {
	if !this.commited {
		return this.tx.Rollback()
	}
	return nil
}

func (this *PgTxHolder) Commit() error {
	err := this.tx.Commit()
	if err == nil {
		this.commited = true
	}
	return err
}

func (this *PgTxHolder) performUpdates(query string, mapper func(entity interface{}) ([]interface{}, error), values ...interface{}) error {
	stmt, err := this.tx.Prepare(query)
	if err != nil {
		log.Errorf("Failed to prepare query %s: %v", query, err)
		return err
	}
	for _, entity := range values {
		values, err := mapper(entity)
		if err != nil {
			log.Errorf("Can not update %v", err)
			return err
		}
		_, err = stmt.Exec(values...)
		if err != nil {
			log.Errorf("Can not update %v", err)
			return err
		}
	}

	log.Debug("Update completed. Commit.")
	return stmt.Close()
}

// Temporary method before issue #111 implemented
func (this *PostgresStorage) PerformUpdatesWithinTxOptionally(tx interface{}, query string, mapper func(entity interface{}) ([]interface{}, error), values ...interface{}) error {
	if tx != nil {
		txHolder, ok := tx.(PgTxHolder)
		if !ok {
			return fmt.Errorf("Unsupported tx type: %v", reflect.TypeOf(tx))
		}
		return txHolder.performUpdates(query, mapper, values...)
	}
	return this.PerformUpdates(query, mapper, values...)
}
