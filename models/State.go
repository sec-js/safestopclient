package models

import (
	"github.com/schoolwheels/safestopclient/database"
	"fmt"
)

type State struct {
	*ModelBase
	Id	int `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Abbreviation string `json:"abbreviation" db:"abbreviation"`
	CountryId string `json:"country_id" db:"country_id"`
}

func StateForAbbreviation(abbreviation string) *State {
	query := "select id, name, abbreviation, country_id from states where abbreviation = $1 limit 1;"
	row := database.GetDB().QueryRowx(query, abbreviation)
	if row == nil {
		return nil
	} else {
		r := State{}
		err := row.StructScan(&r)
		if err != nil {
			fmt.Print(err)
		}
		return &r
	}
}
