package dao

import (
	"database/sql"
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/and-hom/wwmap/lib/config"
	"io"
	"reflect"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(c config.Db) PostgresStorage {
	db, err := sql.Open("postgres", c.ConnString)
	if err != nil {
		log.Fatalf("Can not connect to postgres: %v", err)
	}
	db.SetConnMaxLifetime(c.MaxConnLifetime)
	db.SetMaxOpenConns(c.MaxOpenConn)
	db.SetMaxIdleConns(c.MaxIddleConn)

	return PostgresStorage{
		db: db,
	}
}

func NewPostgresStorageForDb(db *sql.DB) PostgresStorage {
	return PostgresStorage{db}
}

func nullIf0(x int64) sql.NullInt64 {
	if x == 0 {
		return sql.NullInt64{
			Int64: 0,
			Valid: false,
		}
	}
	return sql.NullInt64{
		Int64: x,
		Valid: true,
	}
}

func getOrElse(val sql.NullInt64, _default int64) int64 {
	if val.Valid {
		return val.Int64
	} else {
		return _default
	}
}

func (this *PostgresStorage) doFindAndReturn(query string, callback interface{}, args ...interface{}) (interface{}, bool, error) {
	rows, err := this.db.Query(query, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	funcValue := reflect.ValueOf(callback)

	for rows.Next() {
		val := funcValue.Call([]reflect.Value{reflect.ValueOf(rows)})
		if val[1].Interface() == nil {
			return val[0].Interface(), true, nil
		} else {
			return nil, false, val[1].Interface().(error)
		}
	}
	return nil, false, nil
}

func (this *PostgresStorage) doFindList(query string, callback interface{}, args ...interface{}) (interface{}, error) {
	rows, err := this.db.Query(query, args...)
	if err != nil {
		return []interface{}{}, err
	}
	defer rows.Close()

	funcValue := reflect.ValueOf(callback)
	returnType := funcValue.Type().Out(0)
	var result = reflect.MakeSlice(reflect.SliceOf(returnType), 0, 0)

	var lastErr error = nil
	for rows.Next() {
		val := funcValue.Call([]reflect.Value{reflect.ValueOf(rows)})
		if val[1].Interface() == nil {
			result = reflect.Append(result, val[0])
		} else {
			log.Error(val[1])
			lastErr = (val[1]).Interface().(error)
			break
		}
	}
	return result.Interface(), lastErr
}

func (this *PostgresStorage) forEach(query string, callback interface{}, args ...interface{}) error {
	rows, err := this.db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	funcValue := reflect.ValueOf(callback)

	for rows.Next() {
		val := funcValue.Call([]reflect.Value{reflect.ValueOf(rows)})
		if val[0].Interface() != nil {
			return val[0].Interface().(error)
		}
	}
	return nil
}

// Deprecated: use updateReturningId
func (this *PostgresStorage) insertReturningId(query string, args ...interface{}) (int64, error) {
	tx, err := this.db.Begin()
	if err != nil {
		return -1, err
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		return -1, err
	}
	rows, err := stmt.Query(args...)
	if err != nil {
		return -1, err
	}

	lastId := int64(-1)
	for rows.Next() {
		rows.Scan(&lastId)
	}

	err = rows.Close()
	if err != nil {
		return -1, err
	}
	err = stmt.Close()
	if err != nil {
		return -1, err
	}
	err = tx.Commit()
	if err != nil {
		return -1, err
	}
	if lastId < 0 {
		return -1, errors.New("Not inserted")
	}
	return lastId, nil
}

func (this *PostgresStorage) updateReturningId(query string,
	mapper func(entity interface{}) ([]interface{}, error),
	failOnEmptyResult bool, values ...interface{}) ([]int64, error) {

	rows, err := this.updateReturningColumns(query, mapper, failOnEmptyResult, values...)
	if err != nil {
		return []int64{}, err
	}
	result := this.getFirstColumnAsInt64(rows)
	return result, nil
}

func (this *PostgresStorage) updateReturningColumns(query string,
	mapper func(entity interface{}) ([]interface{}, error),
	failOnEmptyResult bool, values ...interface{}) ([][]interface{}, error) {
	return this.updateReturningColumnsWithinTxOptionally(nil, query, mapper, failOnEmptyResult, values...)
}

func (this *PostgresStorage) getFirstColumnAsInt64(rows [][]interface{}) []int64 {
	result := make([]int64, len(rows))
	for i, row := range rows {
		result[i] = *row[0].(*int64)
	}
	return result
}

func (this *PostgresStorage) updateReturningColumnsWithinTxOptionally(txHolder interface{},
	query string,
	mapper func(entity interface{}) ([]interface{}, error),
	failOnEmptyResult bool, values ...interface{}) ([][]interface{}, error) {
	var tx *sql.Tx
	var err error

	if txHolder == nil {
		tx, err = this.db.Begin()
		if err != nil {
			return [][]interface{}{}, err
		}
		defer func() {
			if err != nil {
				err = tx.Rollback()
				log.Error("Can't rollback: ", err)
				return
			}
			err = tx.Commit()
			if err != nil {
				log.Error("Can't commit: ", err)
			}
		}()
	} else {
		tx = txHolder.(PgTxHolder).tx
	}
	return this.updateReturningColumnsInternal(tx, query, mapper, failOnEmptyResult, values...)
}

func (this *PostgresStorage) updateReturningColumnsInternal(tx *sql.Tx, query string,
	mapper func(entity interface{}) ([]interface{}, error),
	failOnEmptyResult bool, values ...interface{}) ([][]interface{}, error) {

	stmt, err := tx.Prepare(query)
	if err != nil {
		return [][]interface{}{}, err
	}
	defer deferCloser(stmt)

	result := make([][]interface{}, 0, len(values))
	for _, value := range values {
		args, err := mapper(value)
		if err != nil {
			return [][]interface{}{}, err
		}
		rows, err := stmt.Query(args...)
		if err != nil {
			return [][]interface{}{}, err
		}

		colTypes, err := rows.ColumnTypes()
		if err != nil {
			return [][]interface{}{}, err
		}
		if rows.Next() {
			resultValue := make([]interface{}, len(colTypes))
			for i, t := range colTypes {
				resultValue[i] = reflect.New(t.ScanType()).Interface()
			}
			err = rows.Scan(resultValue...)
			if err != nil {
				return [][]interface{}{}, err
			}
			result = append(result, resultValue)
		} else if failOnEmptyResult {
			return [][]interface{}{}, fmt.Errorf("Value is not inserted: %v+\n %s", args, query)
		}
		deferCloser(rows)
	}
	if failOnEmptyResult && len(result) != len(values) {
		return [][]interface{}{}, fmt.Errorf("Some values are not inserted. Values len is %d and result len is %d: %v+\n %s",
			len(values), len(result), result, query)
	}
	return result, nil
}

func (this *PostgresStorage) performUpdates(query string, mapper func(entity interface{}) ([]interface{}, error), values ...interface{}) error {
	return this.WithinTx(func(tx interface{}) error {
		txHolder := tx.(PgTxHolder)
		return (&txHolder).performUpdates(query, mapper, values...)
	})
}

func deferCloser(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		log.Error("Can't close: ", err)
	}
}
