package dao

import "database/sql"

func IdMapper(_id interface{}) ([]interface{}, error) {
	return []interface{}{_id}, nil
}

func ArrayMapper(arr interface{}) ([]interface{}, error) {
	return arr.([]interface{}), nil
}

func StringColumnMapper(rows *sql.Rows) (string, error) {
	i := ""
	err := rows.Scan(&i)
	return i, err
}

func BoolColumnMapper(rows *sql.Rows) (bool, error) {
	i := false
	err := rows.Scan(&i)
	return i, err
}

func IntColumnMapper(rows *sql.Rows) (int, error) {
	i := 0
	err := rows.Scan(&i)
	return i, err
}

func Int64ColumnMapper(rows *sql.Rows) (int64, error) {
	var i int64
	err := rows.Scan(&i)
	return i, err
}
