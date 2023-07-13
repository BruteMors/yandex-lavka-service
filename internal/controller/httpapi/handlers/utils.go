package handlers

import (
	"github.com/go-playground/validator/v10"
	"regexp"
)

var regexpTimeIntervalValidate *regexp.Regexp
var TimeIntervalLength = len("HH:MM-HH:MM")

func init() {
	regexpTimeIntervalValidate = regexp.MustCompile("^([0-1]?[0-9]|2[0-3]):[0-5][0-9]-([0-1]?[0-9]|2[0-3]):[0-5][0-9]$")
}

// ValidateTimeInterval - custom validation func for time format like HH:MM-HH:MM
func ValidateTimeInterval(fl validator.FieldLevel) bool {
	TimeInterval := fl.Field().String()

	if len(TimeInterval) != TimeIntervalLength {
		return false
	}
	result := regexpTimeIntervalValidate.MatchString(TimeInterval)
	return result
}
