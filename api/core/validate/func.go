package validate

import (
	"github.com/go-playground/validator/v10"
	"regexp"
)

// 自定义验证方法：检查字段是否包含宏参数格式
func IsMacroParam(fl validator.FieldLevel) bool {
	field := fl.Field().String()
	macroRegex := regexp.MustCompile(`^__\w+__$`)
	return !macroRegex.MatchString(field) // 若为宏格式，则验证失败
}
