package common

import (
	"github.com/go-playground/validator/v10"
	"reflect"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New()
	_ = Validate.RegisterValidation("ValidateEmbeddingInput", validateEmbeddingInput)
}

func validateEmbeddingInput(fl validator.FieldLevel) bool {
	v := fl.Field()
	var check func(v reflect.Value, mustBe reflect.Kind) bool
	check = func(v reflect.Value, mustBe reflect.Kind) bool {
		if mustBe != reflect.Invalid && v.Kind() != mustBe {
			return false
		}
		switch v.Kind() {
		case reflect.String:
			return true
		case reflect.Array, reflect.Slice:
			if v.Len() == 0 {
				return false
			}
			for i := 0; i < v.Len(); i++ {
				checkResult := check(v.Index(i), reflect.String)
				if v.Index(i).Kind() == reflect.Interface || v.Index(i).Kind() == reflect.Ptr {
					checkResult = checkResult || check(v.Index(i).Elem(), reflect.String)
				}
				if !checkResult {
					return false
				}
			}
		default:
			return false
		}
		return true
	}
	return check(v, reflect.Invalid)
}
