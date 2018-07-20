package i18n

import (
	"github.com/qor/i18n"
	"github.com/qor/i18n/backends/yaml"
)

var i18nInstance *i18n.I18n

func GetI18n() *i18n.I18n {
	if i18nInstance == nil {
		i18nInstance = i18n.New(
			yaml.New("./config/locales"), // load translations from the YAML files in directory `config/locales`
		)
	}

	return i18nInstance
}

