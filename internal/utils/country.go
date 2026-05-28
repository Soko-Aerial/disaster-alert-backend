package utils

import (
	"strings"

	"github.com/biter777/countries"
)

func NormalizeCountryCode(value string) string {
	value = strings.TrimSpace(value)

	if value == "" {
		return ""
	}

	country := countries.ByName(value)
	if country != countries.Unknown {
		return country.Alpha2()
	}

	country = countries.ByName(strings.ToUpper(value))
	if country != countries.Unknown {
		return country.Alpha2()
	}

	country = countries.ByName(strings.ToLower(value))
	if country != countries.Unknown {
		return country.Alpha2()
	}

	country = countries.ByName(strings.Title(strings.ToLower(value)))
	if country != countries.Unknown {
		return country.Alpha2()
	}

	return strings.ToUpper(value)
}