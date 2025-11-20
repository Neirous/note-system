package common

//定义统一响应结构体
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

//快捷响应函数
//成功响应
func Success(data interface{}) Response {
	return Response{
		0,
		"success",
		data,
	}
}

func Fail(msg string) Response {
	return Response{
		1,
		msg,
		nil,
	}
}
