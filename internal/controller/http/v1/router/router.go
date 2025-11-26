package router

import "github.com/ducklawrence05/go-test-backend-api/internal/controller/http/v1/router/user"

type RouterGroup struct {
	User user.RouterGroup
}

var RouterGroupApp = new(RouterGroup)
