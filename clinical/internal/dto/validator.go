package dto

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"regexp"
)

var Validate *validator.Validate
var matched = regexp.MustCompile(`^[A-Z0-9-]+$`)

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())

	// Optional: register custom tag name extractor (better error messages)
	Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	Validate.RegisterValidation("code", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if value == "" {
			return false
		}
		return matched.MatchString(value)
	})
}
