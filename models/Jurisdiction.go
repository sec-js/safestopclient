package models

import (
	"github.com/schoolwheels/safestopclient/database"
	"fmt"
		)


type JurisdictionOptions struct {
	AuthInfo AuthInfo `json:"auth_info"`
	Error Error `json:"error"`
	Jurisdictions []JurisdictionOption `json:"jurisdictions"`
}

type JurisdictionOption struct {
	*ModelBase
	Id	int `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

func AvailableJurisdictionsForState(resp *JurisdictionOptions, state_id int) {

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
		resp.Error.Msg = err.Error()
	}

	if rows != nil {
		for i := 0; rows.Next(); i++ {
			r := JurisdictionOption{}
			err := rows.StructScan(&r)
			if err != nil {
				fmt.Print(err)
			}
			resp.Jurisdictions = append(resp.Jurisdictions, r)
		}
	}
}