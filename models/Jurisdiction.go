package models

import (
	"github.com/schoolwheels/safestopclient/database"
	"fmt"
	"log"
)


type JurisdictionOptions struct {
	Jurisdictions []JurisdictionOption `json:"jurisdictions"`
}

type JurisdictionOption struct {
	*ModelBase
	Id	int `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

func AvailableJurisdictionsForState(state_id int) *JurisdictionOptions {
	j := JurisdictionOptions{}

	query := `
select a.id, a.name 
from jurisdictions a
join products b on b.jurisdiction_id = a.id
join time_zones c on c.id = a.time_zone_id
where state_id = $1
and b.active = 't'
and b.product_type = 'ss'
and a.active = 't'
and b.availability_start_date <= now() at time zone c.postgresql_name
and b.availability_end_date >= now() at time zone c.postgresql_name
order by a.name
`
	rows, err := database.GetDB().Queryx(query, state_id)
	if err != nil {
		fmt.Print(err)
	}

	if rows != nil {
		for i := 0; rows.Next(); i++ {
			r := JurisdictionOption{}
			err := rows.StructScan(&r)
			if err != nil {
				fmt.Print(err)
			}
			j.Jurisdictions = append(j.Jurisdictions, r)
		}
	}

	return &j
}

func JurisdictionCountForUser(u *User, pg *PermissionGroups) int {

	query := ``
	jurisdiction_count := 0

	if((u.SuperAdmin == true) || HasAnyPermissionGroups([]string{ pg.Admin}, u.PermissionGroups)) {
		query = `
select count(*) 
from jurisdictions 
where active = true;
`
		err := database.GetDB().QueryRow(query).Scan(&jurisdiction_count)
		if err != nil {
			log.Fatal(err)
		}

	} else if (HasAnyPermissionGroups([]string{ pg.License_1, pg.License_2, pg.License_3, pg.License_4}, u.PermissionGroups)){
		query = `
select count(*) from jurisdictions a 
join jurisdictional_restrictions b on a.id = b.jurisdiction_id
where b.user_id = $1
`

		err := database.GetDB().QueryRow(query, u.Id).Scan(&jurisdiction_count)
		if err != nil {
			log.Fatal(err)
		}

	} else {
	query = `
select count(distinct id) from (

select a.id 
from jurisdictions a
join products b on b.jurisdiction_id = a.id
join subscriptions c on b.id = c.product_id
join subscription_sub_accounts d on c.id = d.subscription_id
join users e on d.user_id = e.id
join time_zones f on f.id = a.time_zone_id
where c.start_date <= (now() at time zone f.postgresql_name)::date
and c.end_date >= (now() at time zone f.postgresql_name)::date
and c.active = 't'
and b.product_type = 'ss'
and e.id = $1

union all

select a.id
from jurisdictions a
join products b on b.jurisdiction_id = a.id
join subscriptions c on b.id = c.product_id
join users d on c.user_id = d.id
join time_zones e on e.id = a.time_zone_id
where c.start_date <= (now() at time zone e.postgresql_name)::date
and c.end_date >= (now() at time zone e.postgresql_name)::date
and c.active = 't'
and b.product_type = 'ss'
and d.id = $1

) z
`
		err := database.GetDB().QueryRow(query, u.Id).Scan(&jurisdiction_count)
		if err != nil {
			log.Fatal(err)
		}
	}

	return jurisdiction_count
}


func SchoolCodeExists(school_code string, jurisdiction_id int) bool {
	code_ct := 0
	query := `
select count(*)
from jurisdictions 
where gate_key = $1
and id = $2
and active = true
`
	row := database.GetDB().QueryRowx(query, school_code, jurisdiction_id)
	if row == nil {
		return true
	} else {
		err := row.Scan(&code_ct)
		if err != nil {
			return true
		}
		return (code_ct > 0)
	}
}