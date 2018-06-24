package controllers

import (
	"github.com/XMatrixStudio/Coffee/App/services"
	"github.com/kataras/iris"
	"github.com/kataras/iris/sessions"
)

// ContentController 内容
type ContentController struct {
	Ctx     iris.Context
	Service services.ContentService
	Session *sessions.Session
}
