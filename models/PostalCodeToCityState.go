package models

import (
	"github.com/schoolwheels/safestopclient/database"
		"log"
)

type PostalCodeReference struct {
	*ModelBase
	Id	int `json:"id" db:"id"`
	PostalCode string `json:"postal_code" db:"postal_code"`
	StateCode string `json:"state_code" db:"state_code"`
	City string `json:"city" db:"city"`
	Latitude string `json:"latitude" db:"latitude"`
	Longitude string `json:"longitude" db:"longitude"`
}

func PostalCodeReferenceForPostalCode(postal_code string) *PostalCodeReference {
	query := "select id, postal_code, state_code, city, latitude, longitude from postal_code_to_city_states where postal_code = $1 limit 1;"
	row := database.GetDB().QueryRowx(query, postal_code)

	r := PostalCodeReference{}
	err := row.StructScan(&r)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &r
}
