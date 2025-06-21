package validate

import (
	"errors"
	"github.com/go-playground/locales"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
)

type Validate struct {
	validate *validator.Validate
	trans    ut.Translator
}

// RegisterValidation 注册自定义验证方法
// tag 为验证标签名称，fn 为自定义验证方法
func (v *Validate) RegisterValidation(tag string, fn validator.Func) error {
	return v.validate.RegisterValidation(tag, fn)
}

// InitValidates 初始化相关可用数据
func (v *Validate) InitValidates(localTrans locales.Translator, local string) {
	uni := ut.New(localTrans, localTrans)
	// this is usually know or extracted from http 'Accept-Language' header
	// also see uni.FindTranslator(...)
	v.trans, _ = uni.GetTranslator(local)

	v.validate = validator.New()

	err := zh_translations.RegisterDefaultTranslations(v.validate, v.trans)
	if err != nil {
		panic(err)
	}
	// 注册自定义验证 "nomacro"
	err = v.validate.RegisterValidation("nomacro", IsMacroParam)
	if err != nil {
		return
	}
}

// HandleError 处理错误
// r 为验证的赋值模型
// m 为自定义错误消息
func (v *Validate) HandleError(r interface{}, m map[string]string) error {
	err := v.validate.Struct(r)
	if err != nil {
		errs := err.(validator.ValidationErrors)
		for _, e := range errs {
			if _, ok := m[e.Field()+"."+e.Tag()]; ok {
				return errors.New(m[e.Field()+"."+e.Tag()])
			} else {
				tranStr := e.Translate(v.trans)
				return errors.New(tranStr)
			}
		}
	}

	return nil
}

// New 执行验证
// r 为验证的赋值模型
// m 为自定义错误消息
func New(r interface{}, m map[string]string, localTrans locales.Translator, local string) error {
	v := Validate{}
	v.InitValidates(localTrans, local)
	err := v.HandleError(r, m)
	return err
}

// new validator
func Run(r interface{}, m map[string]string) error {
	// 这里默认是中文，可以根据需求进行修改，就没有做过多的封装了
	err := New(r, m, zh.New(), "zh")
	return err
}
