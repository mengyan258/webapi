package minimal

import (
	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/core/eumLogLevel"
	"github.com/farseer-go/fs/flog"
	"github.com/farseer-go/fs/types"
	"github.com/farseer-go/webapi/context"
	"github.com/farseer-go/webapi/middleware"
	"reflect"
)

// Register 注册单个Api
func Register(area string, method string, route string, actionFunc any, paramNames ...string) context.HttpRoute {
	actionType := reflect.TypeOf(actionFunc)
	param := types.GetInParam(actionType)

	// 如果设置了方法的入参（多参数），则需要全部设置
	if len(paramNames) > 0 && len(paramNames) != len(param) {
		panic(flog.Errorf("注册minimalApi失败：%s函数入参个数设置与%s不匹配", flog.Colors[eumLogLevel.Error](actionType.String()), flog.Colors[eumLogLevel.Error](paramNames)))
	}

	lstRequestParamType := collections.NewList(param...)
	lstResponseParamType := collections.NewList(types.GetOutParam(actionType)...)

	// 添加到路由表
	return context.HttpRoute{
		RouteUrl:            area + route,
		Action:              actionFunc,
		Method:              method,
		RequestParamType:    lstRequestParamType,
		ResponseBodyType:    lstResponseParamType,
		ParamNames:          collections.NewList(paramNames...),
		RequestParamIsModel: types.IsDtoModel(lstRequestParamType.ToArray()),
		ResponseBodyIsModel: types.IsDtoModel(lstResponseParamType.ToArray()),
		HttpMiddleware:      &middleware.Http{},
		HandleMiddleware:    &HandleMiddleware{},
		IsGoBasicType:       types.IsGoBasicType(lstResponseParamType.First()),
	}
}
