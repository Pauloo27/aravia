package aravia

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Field   string `json:"field"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

var (
	engine *validator.Validate
)

func init() {
	engine = validator.New()
	engine.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

}

func Validate(a interface{}) []*ValidationError {
	var errs []*ValidationError

	rawErrs := engine.Struct(a)

	if rawErrs == nil {
		return nil
	}

	for _, err := range rawErrs.(validator.ValidationErrors) {
		errs = append(errs, &ValidationError{
			Message: err.Error(),
			Field:   err.Field(),
			Type:    err.Tag(),
		})
	}

	return errs
}
