package validatorx

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// Init 全局初始化校验器，注册自定义tag、自定义错误翻译
func Init() {
	validate = validator.New()
	// 注册获取结构体json标签，错误提示显示json字段名而非结构体字段名
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// Get 获取全局校验实例
func Get() *validator.Validate {
	return validate
}

// TranslateErr 将校验错误转为中文可读字符串
func TranslateErr(err error) string {
	if err == nil {
		return ""
	}
	var msgs []string
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err.Error()
	}
	for _, e := range errs {
		switch e.Tag() {
		case "required":
			msgs = append(msgs, "参数"+e.Field()+"不能为空")
		case "email":
			msgs = append(msgs, "参数"+e.Field()+"格式错误，请输入合法邮箱")
		case "min":
			msgs = append(msgs, "参数"+e.Field()+"最小值为"+e.Param())
		case "max":
			msgs = append(msgs, "参数"+e.Field()+"最大值为"+e.Param())
		case "len":
			msgs = append(msgs, "参数"+e.Field()+"长度必须为"+e.Param())
		case "gte":
			msgs = append(msgs, "参数"+e.Field()+"不能小于"+e.Param())
		case "lte":
			msgs = append(msgs, "参数"+e.Field()+"不能大于"+e.Param())
		default:
			msgs = append(msgs, "参数"+e.Field()+"校验不通过，规则："+e.Tag())
		}
	}
	return strings.Join(msgs, ";")
}
