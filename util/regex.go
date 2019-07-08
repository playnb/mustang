package util

import "regexp"

const (
	regular = "^1([34578][0-9])\\d{8}$"
)

func ValidateMobileNum(mobileNum string) bool {
	reg := regexp.MustCompile(regular)
	return reg.MatchString(mobileNum)
}
