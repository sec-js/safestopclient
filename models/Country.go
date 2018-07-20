package models

import (
	"fmt"
	"log"
	"github.com/schoolwheels/safestopclient/database"
)

type Country struct {
	*ModelBase
	Id int `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Code string `json:"code" db:"code"`
}

func FindCountryById(id int) *Country {

	queryFindCountry := "select * from countries where id = $1;"
	row := database.GetDB().QueryRowx(queryFindCountry, id)
	if row == nil {
		return nil
	} else {
		c := Country{}
		err := row.StructScan(&c)
		if err != nil {
			fmt.Print(err)
		}
		return &c
	}
}

func FindCountryByName(name string) *Country {

	queryFindCountry := "select * from countries where name = $1;"
	row := database.GetDB().QueryRowx(queryFindCountry, name)
	if row == nil {
		return nil
	} else {
		c := Country{}
		err := row.StructScan(&c)
		if err != nil {
			fmt.Print(err)
		}
		return &c
	}
}

func FindCountryByCode(code string) *Country {

	queryFindCountry := "select * from countries where code = $1;"
	row := database.GetDB().QueryRowx(queryFindCountry, code)
	if row == nil {
		return nil
	} else {
		c := Country{}
		err := row.StructScan(&c)
		if err != nil {
			fmt.Print(err)
		}
		return &c
	}
}

func AllCountries() []Country {

	queryAllCountries := "select * from countries;"
	rows, err := database.GetDB().Queryx(queryAllCountries)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var countries []Country
	for rows.Next() {
		c := Country{}
		err := rows.StructScan(&c)
		if err != nil {
			log.Fatal(err)
		}

		countries = append(countries, c)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	return countries
}

func GetCountryByName(name string) *Country {
	row := database.GetDB().QueryRowx(`SELECT * FROM countries WHERE name = $1`, name)

	if row == nil {
		return nil
	} else {
		c := Country{}
		err := row.StructScan(&c)
		if err != nil {
			fmt.Print(err)
		}
		return &c
	}
}
