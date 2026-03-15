package service

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/knnedy/projectflow/internal/domain"
)

func newValidator() (*validator.Validate, ut.Translator) {
	enLocale := en.New()
	uni := ut.New(enLocale, enLocale)

	trans, _ := uni.GetTranslator("en")

	validate := validator.New()

	// register human readable field names from json tags
	// so errors say "email" not "Email"
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// register english translations
	en_translations.RegisterDefaultTranslations(validate, trans)

	return validate, trans
}

func formatValidationError(err error, trans ut.Translator) error {
	validationErrors := err.(validator.ValidationErrors)
	firstErr := validationErrors[0]
	return &domain.ValidationError{
		Field:   firstErr.Field(),
		Message: firstErr.Translate(trans),
	}
}
