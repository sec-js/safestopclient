package models

import (
	"github.com/schoolwheels/safestopclient/database"
	"log"
)

type Subscriptions struct {
	Subscriptions []Subscription
}

type Subscription struct {
	Id int `json:"id" db:"id"`
	ProductId int `json:"product_id" db:"product_id"`
	JurisdictionId int `json:"jurisdiction_id" db:"jurisdiction_id"`
	JurisdictionName string `json:"jurisdiction_name" db:"jurisdiction_name"`
	RegistrationType string `json:"registration_type" db:"registration_type"`
	StartDate string `json:"start_date" db:"start_date"`
	EndDate string `json:"end_date" db:"end_date"`
}


func SubscriptionsForUser(user *User) *Subscriptions {
	subs := Subscriptions{}

	query := `
select 
c.id as id,
a.id as jurisdiction_id, 
a.name as jurisdiction_name, 
f.name as registration_type
from jurisdictions a
join products b on b.jurisdiction_id = a.id
join subscriptions c on b.id = c.product_id
join users d on c.user_id = d.id
join time_zones e on e.id = a.time_zone_id
join safe_stop_registration_types f on f.id = a.safe_stop_registration_type_id
where c.start_date <= (now() at time zone e.postgresql_name)::date
and c.end_date >= (now() at time zone e.postgresql_name)::date
and c.active = 't'
and b.product_type = 'ss'
and d.id = $1
order by jurisdiction_name
`
	rows, err := database.GetDB().Queryx(query, user.Id)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	for rows.Next() {
		sub := Subscription{}
		err = rows.StructScan(&sub)
		if err != nil {
			log.Println(err.Error())
			return nil
		}
		subs.Subscriptions = append(subs.Subscriptions, sub)
	}

	if len(subs.Subscriptions) > 0 {
		return &subs
	} else {
		return nil
	}
}