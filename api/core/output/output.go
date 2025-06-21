package output

import "github.com/kataras/iris/v12"

/**
 * 封装的输出类，返回指定的格式信息
 */

const (
	STATUS_SUCCESS = 200 //成功
	STATUS_ERROR   = 400 //失败
)

const MSG_OK = "ok"

// Output 输出结构体
type Output struct {
	Code int
	Msg  string
	Data interface{}
	Ctx  iris.Context
}

// Success 输出成功
func (op *Output) Success(data interface{}) {
	if op.Code == 0 {
		op.Code = STATUS_SUCCESS
	}
	op.Msg = MSG_OK
	op.Data = data

	op.OutJson()
}

// Msg 输出标准错误
func (op *Output) Error(message string) {
	if op.Code == 0 {
		op.Code = STATUS_ERROR
	}
	op.Msg = message

	op.OutJson()
}

func (op *Output) OutJson() {
	_ = op.Ctx.JSON(map[string]interface{}{
		"code":    op.Code,
		"message": op.Msg,
		"data":    op.Data,
	})
}
