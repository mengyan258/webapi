package context

import (
	"bytes"
	"encoding/json"
	"github.com/farseer-go/collections"
	"github.com/farseer-go/fs/parse"
	"github.com/farseer-go/fs/types"
	"net/http"
	"reflect"
	"strings"
)

type HttpContext struct {
	HttpRequest      HttpRequest
	HttpResponse     HttpResponse
	HttpHeader       collections.Dictionary[string, string]
	HttpURL          HttpURL
	Method           string
	ContentLength    int64
	Close            bool
	TransferEncoding []string
	ContentType      string
}

// NewHttpContext 初始化上下文
func NewHttpContext(r *http.Request) HttpContext {
	// Body
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r.Body)
	var httpContext = HttpContext{
		HttpRequest: HttpRequest{
			Body:       r.Body,
			BodyString: buf.String(),
			BodyBytes:  buf.Bytes(),
		},
		HttpResponse: HttpResponse{},
		HttpHeader:   collections.NewDictionary[string, string](),
		HttpURL: HttpURL{
			Path:        r.URL.Path,
			RemoteAddr:  r.RemoteAddr,
			Host:        r.Host,
			Proto:       r.Proto,
			RequestURI:  r.RequestURI,
			QueryString: r.URL.RawQuery,
			Query:       collections.NewDictionary[string, string](),
		},
		Method:           r.Method,
		ContentLength:    r.ContentLength,
		Close:            r.Close,
		TransferEncoding: r.TransferEncoding,
		ContentType:      "",
	}

	switch httpContext.Method {
	case "GET":
		_ = r.ParseForm()
		httpContext.HttpRequest.ParseQuery(r.Form)
	default:
		httpContext.HttpRequest.ParseForm()
	}

	// httpURL
	for k, v := range r.URL.Query() {
		httpContext.HttpURL.Query.Add(k, strings.Join(v, ";"))
	}

	if r.TLS == nil {
		httpContext.HttpURL.Url = "http://" + r.Host + r.RequestURI
	} else {
		httpContext.HttpURL.Url = "https://" + r.Host + r.RequestURI
	}

	// header
	for k, v := range r.Header {
		httpContext.HttpHeader.Add(k, strings.Join(v, ";"))
	}

	// ContentType
	for _, contentType := range strings.Split(httpContext.HttpHeader.GetValue("Content-Type"), ";") {
		if strings.Contains(contentType, "/") {
			httpContext.ContentType = contentType
		}
	}
	return httpContext
}

// GetRequestParam 根据method映射入参
func (httpContext *HttpContext) GetRequestParam(requestParamType collections.List[reflect.Type], paramNames collections.List[string]) []reflect.Value {
	// 没有入参时，忽略request.body
	if requestParamType.Count() == 0 {
		return []reflect.Value{}
	}

	// application/json
	switch httpContext.ContentType {
	case "application/json":
		return httpContext.contentTypeJson(requestParamType, paramNames)
	case "": // GET
		return httpContext.query(requestParamType, paramNames)
	default: //case "application/x-www-form-urlencoded", "multipart/form-data":
		return httpContext.formUrlencoded(requestParamType, paramNames)
	}

	//lstParams := collections.NewList[reflect.Value]()
	//for _, paramType := range requestParamType.ToArray() {
	//	lstParams.Add(reflect.New(paramType).Elem())
	//}
	//
	//return lstParams.ToArray()
}

// application/json
func (httpContext *HttpContext) contentTypeJson(requestParamType collections.List[reflect.Type], paramNames collections.List[string]) []reflect.Value {
	// dto
	firstParamType := requestParamType.First() // 先取第一个参数
	isCollections := types.IsCollections(firstParamType)
	isStruct := firstParamType.Kind() == reflect.Struct
	if requestParamType.Count() == 1 && !isCollections && isStruct {
		val := reflect.New(firstParamType).Interface()
		json.Unmarshal(httpContext.HttpRequest.BodyBytes, val)
		return []reflect.Value{reflect.ValueOf(val).Elem()}
	}

	// 多参数
	mapVal := httpContext.HttpRequest.JsonToMap()
	return httpContext.mapToParams(requestParamType, paramNames, mapVal)
}

// application/x-www-form-urlencoded
func (httpContext *HttpContext) formUrlencoded(requestParamType collections.List[reflect.Type], paramNames collections.List[string]) []reflect.Value {
	// 多参数
	return httpContext.mapToParams(requestParamType, paramNames, httpContext.HttpRequest.Form)
}

// query
func (httpContext *HttpContext) query(requestParamType collections.List[reflect.Type], paramNames collections.List[string]) []reflect.Value {
	// 多参数
	return httpContext.mapToParams(requestParamType, paramNames, httpContext.HttpRequest.Query)
}

// 将map转成入参值
func (httpContext *HttpContext) mapToParams(requestParamType collections.List[reflect.Type], paramNames collections.List[string], mapVal map[string]any) []reflect.Value {
	lstParams := make([]reflect.Value, requestParamType.Count())
	for i := 0; i < requestParamType.Count(); i++ {
		defVal := reflect.New(requestParamType.Index(i)).Elem().Interface()
		if paramNames.Count() > i {
			paramName := strings.ToLower(paramNames.Index(i))
			paramVal, _ := mapVal[paramName]
			defVal = parse.Convert(paramVal, defVal)
		}
		lstParams[i] = reflect.ValueOf(defVal)
	}
	return lstParams
}

// InitResponse 初始化返回报文
func (httpContext *HttpContext) InitResponse(returnVals []reflect.Value) {
	// 没有返回值，则不响应
	if len(returnVals) == 0 {
		httpContext.HttpResponse.BodyBytes = []byte{}
		httpContext.HttpResponse.BodyString = ""
		return
	}
	firstParamType := types.GetRealType(returnVals[0])
	isCollections := types.IsCollections(firstParamType)
	isStruct := firstParamType.Kind() == reflect.Struct

	if len(returnVals) == 1 {
		responseBody := returnVals[0].Interface()

		if !isCollections && isStruct { // dto
			httpContext.HttpResponse.BodyBytes, _ = json.Marshal(responseBody)
			httpContext.HttpResponse.BodyString = string(httpContext.HttpResponse.BodyBytes)
		} else { // 基本类型直接转string
			httpContext.HttpResponse.BodyString = parse.Convert(responseBody, "")
			httpContext.HttpResponse.BodyBytes = []byte(httpContext.HttpResponse.BodyString)
		}
	}

	if len(returnVals) > 1 {
		responseBody := returnVals[0].Interface()
		httpContext.HttpResponse.BodyBytes, _ = json.Marshal(responseBody)
		httpContext.HttpResponse.BodyString = string(httpContext.HttpResponse.BodyBytes)
	}
}
