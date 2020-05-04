package dao

import (
	"encoding/json"
	"fmt"
)

type PropertyManager interface {
	GetIntProperty(name string, id int64) (int, error)
	SetIntProperty(name string, id int64, value int) error
	GetBoolProperty(name string, id int64) (bool, error)
	SetBoolProperty(name string, id int64, value bool) error
	GetStringProperty(name string, id int64) (string, error)
	SetStringProperty(name string, id int64, value string) error
	RemoveProperty(name string, id int64) error
}

type PropertyManagerImpl struct {
	dao   *PostgresStorage
	table string
}

func (this PropertyManagerImpl) GetIntProperty(name string, id int64) (int, error) {
	i, found, err := this.getProperty(name, "int", IntColumnMapper, id)
	if err != nil {
		return 0, err
	}
	if !found {
		return 0, nil
	}
	return i.(int), nil
}

func (this PropertyManagerImpl) SetIntProperty(name string, id int64, value int) error {
	return this.setProperty(name, id, value)
}

func (this PropertyManagerImpl) GetBoolProperty(name string, id int64) (bool, error) {
	i, found, err := this.getProperty(name, "bool", BoolColumnMapper, id)
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	return i.(bool), nil
}

func (this PropertyManagerImpl) SetBoolProperty(name string, id int64, value bool) error {
	return this.setProperty(name, id, value)
}

func (this PropertyManagerImpl) GetStringProperty(name string, id int64) (string, error) {
	i, found, err := this.getProperty(name, "varchar", StringColumnMapper, id)
	if err != nil {
		return "", err
	}
	if !found {
		return "", nil
	}
	return i.(string), nil
}

func (this PropertyManagerImpl) SetStringProperty(name string, id int64, value string) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return this.setProperty(name, id, string(b))
}

func (this PropertyManagerImpl) getProperty(name string, sqlType string, rowMapper interface{}, id int64) (interface{}, bool, error) {
	query := fmt.Sprintf("WITH txt_val AS (SELECT (props->>'%s') val FROM %s WHERE id=$1) SELECT val::%s FROM txt_val WHERE val IS NOT NULL",
		name, this.table, sqlType)
	return this.dao.DoFindAndReturn(query, rowMapper, id)
}
func (this PropertyManagerImpl) setProperty(name string, id int64, value interface{}) error {
	query := fmt.Sprintf("UPDATE %s SET props=jsonb_set(props, '{%s}', $2::text::jsonb, true) WHERE id=$1", this.table, name)
	return this.dao.PerformUpdates(query,
		ArrayMapper, []interface{}{id, value})
}
func (this PropertyManagerImpl) RemoveProperty(name string, id int64) error {
	return this.dao.PerformUpdates("UPDATE "+this.table+" SET props=(props - '"+name+"') WHERE id=$1", IdMapper, id)
}
