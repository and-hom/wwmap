package dao

import "fmt"

type EntityType string

const (
	RIVER EntityType = "river"
)

type EntityNotFoundError struct {
	Id         int64
	EntityType EntityType
}

func (this EntityNotFoundError) Error() string {
	return fmt.Sprintf("%s with id %d not found", this.EntityType, this.Id)
}
