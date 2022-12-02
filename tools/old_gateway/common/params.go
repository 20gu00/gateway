package common

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/universal-translator"
	"github.com/pkg/errors"
	"gopkg.in/go-playground/validator.v9"
	"strings"
)

//参数校验
func ValidDefaultParams(ctx *gin.Context, in interface{}) error {
	//ShouldBind检查内容类型以自动选择一个绑定引擎。根据 "Content-Type "头，使用不同的绑定方式。 "application/json" --> JSON绑定  "application/xml" --> XML绑定
	if err := ctx.ShouldBind(in); err != nil {
		return err
	}

	//获取验证器
	valid, err := GetValidator(ctx)
	if err != nil {
		return err
	}

	//获取翻译器
	trans, err := GetTranslation(ctx)
	if err != nil {
		return err
	}

	//验证
	//验证一个结构的暴露字段，并自动验证嵌套结构
	err = valid.Struct(in)
	if err != nil {
		errs := err.(validator.ValidationErrors) //如果不是nil，需要断言错误
		sliceErrs := []string{}
		for _, e := range errs {
			sliceErrs = append(sliceErrs, e.Translate(trans))
		}

		//将字符串切片用,拼接成字符串
		return errors.New(strings.Join(sliceErrs, ","))
	}
	return nil
}

//验证器
func GetValidator(c *gin.Context) (*validator.Validate, error) {
	val, ok := c.Get(ValidatorKey) //从gin的上下文key value获取validator这个键的值
	if !ok {
		return nil, errors.New("未设置validator")
	}

	//对context中的validator的value断言拿到验证器,就是一个结构体
	validator, ok := val.(*validator.Validate) //validator类型都是后者类型,断言成功获取该值,失败就是零值
	if !ok {
		return nil, errors.New("无法获取validator")
	}

	return validator, nil
}

//翻译器
func GetTranslation(c *gin.Context) (ut.Translator, error) {
	trans, ok := c.Get(TranslatorKey)
	if !ok {
		return nil, errors.New("未设置tranclation")
	}

	translator, ok := trans.(ut.Translator)
	if !ok {
		return nil, errors.New("无法获取translation")
	}

	return translator, nil
}
