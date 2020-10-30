package linked_entity

import (
	. "github.com/and-hom/wwmap/backend/handler"
	"github.com/and-hom/wwmap/lib/dao"
	"github.com/and-hom/wwmap/lib/handler"
	"net/http"
)

type linkedEntityHanler struct {
	App
	handler linkedEntityHandler
	base string
}

func (this *linkedEntityHanler) Init() {
	this.Register("/"+this.base, handler.HandlerFunctions{
		Get:  this.handler.List,
		Post: this.ForRoles(this.handler.Upsert, dao.ADMIN, dao.EDITOR),
		Put:  this.ForRoles(this.handler.Upsert, dao.ADMIN, dao.EDITOR),
	})
	this.Register("/"+this.base+"/{id}", handler.HandlerFunctions{
		Delete: this.ForRoles(this.handler.Delete, dao.ADMIN, dao.EDITOR),
		Post:   this.ForRoles(this.handler.Upsert, dao.ADMIN, dao.EDITOR),
		Put:    this.ForRoles(this.handler.Upsert, dao.ADMIN, dao.EDITOR),
		Get:    this.ForRoles(this.handler.Get, dao.ADMIN, dao.EDITOR),
	})
	this.Register("/"+this.base+"/river/{id}", handler.HandlerFunctions{
		Get: this.handler.ByRiver,
	})
}

type linkedEntityHandler interface {
	Create() linkedEntityHanler

	Get(writer http.ResponseWriter, request *http.Request)
	List(writer http.ResponseWriter, request *http.Request)
	Upsert(writer http.ResponseWriter, request *http.Request)
	Delete(writer http.ResponseWriter, request *http.Request)
	ByRiver(writer http.ResponseWriter, request *http.Request)
}
