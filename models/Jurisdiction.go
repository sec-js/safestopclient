package models

import (
	"github.com/schoolwheels/safestopclient/database"
	"fmt"
	"log"
	)


type Jurisdictions struct {
	Jurisdictions []Jurisdiction
}


type Jurisdiction struct {
	Id int `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Phone string `json:"phone" db:"phone"`
	HasLostItemReports bool `json:"has_lost_item_reports" db:"has_lost_item_reports"`
	HasIncidentReports bool `json:"has_incident_reports" db:"has_incident_reports"`
	Active bool `json:"active" db:"active"`
	StudentScanning bool `json:"student_scanning" db:"student_scanning"`
	RegistrationType string `json:"registration_type" db:"registration_type"`
	UseScanCodeMapping bool `json:"use_scan_code_mapping" db:"use_scan_code_mapping"`
	RegisterUrl string `json:"register_url" db:"register_url"`
	ActivateUrl string `json:"activate_url" db:"activate_url"`
	SubAccountLimit int `json:"sub_account_limit" db:"sub_account_limit"`
}

func FindJurisdiction(id int) *Jurisdiction {

	query := `
select a.id, 
coalesce(a.name, '') as name,
coalesce(a.phone, '') as phone ,
coalesce(a.safe_stop_lost_item_reports_active, false) as has_lost_item_reports,
coalesce(a.safe_stop_bullying_reports_active, false) as has_incident_reports,
coalesce(a.active, false) as active,
coalesce(a.student_scanning, false) as student_scanning,
coalesce(a.use_scan_code_mapping, false) as use_scan_code_mapping,
coalesce(c.sub_account_limit, 3) as sub_account_limit,
coalesce(b.name, 'Access Code') as registration_type
from jurisdictions a 
join safe_stop_registration_types b on a.safe_stop_registration_type_id = b.id
join jurisdiction_safe_stop_infos c on c.jurisdiction_id = a.id
where a.id = $1
`
	row := database.GetDB().QueryRowx(query, id)
	if row == nil {
		return nil
	}

	j := Jurisdiction{}
	err := row.StructScan(&j)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &j
}






func AvailableJurisdictionsForState(state_id int, postal_code string) *Jurisdictions {
	j := Jurisdictions{}

	query := `
select a.id, 
a.name,
'/activate/' || a.id || '?postal_code=' || $2 as activate_url,
'/register/' || a.id || '?postal_code=' || $2 as register_url
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
	rows, err := database.GetDB().Queryx(query, state_id, postal_code)
	if err != nil {
		log.Println(err.Error())
		return &j
	}

	for i := 0; rows.Next(); i++ {
		r := Jurisdiction{}
		err := rows.StructScan(&r)
		if err != nil {
			log.Println(err.Error())
			return &j
		}
		j.Jurisdictions = append(j.Jurisdictions, r)
	}

	return &j
}



func JurisdictionCountForUser(u *User, pg *PermissionGroups) int {
	jurisdiction_count := 0
	cj := ClientJurisdictionForUser(u, pg)
	if cj != nil {
		jurisdiction_count = len(cj.Jurisdictions)
	}
	return jurisdiction_count
}

type ClientJurisdictions struct {
	Jurisdictions []ClientJurisdiction
}

type ClientJurisdiction struct {
	Id int `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Phone string `json:"phone" db:"phone"`
	HasLostItemReports bool `json:"has_lost_item_reports" db:"has_lost_item_reports"`
	HasIncidentReports bool `json:"has_incident_reports" db:"has_incident_reports"`
	Active bool `json:"active" db:"active"`
	StudentScanning bool `json:"student_scanning" db:"student_scanning"`
}













func ClientJurisdictionForUser(u *User, pg *PermissionGroups) *ClientJurisdictions {

	client_jurisdictions := ClientJurisdictions{}

	query := ``

	if((u.SuperAdmin == true) || UserHasAnyPermissionGroups([]string{ pg.Admin}, u.PermissionGroups)) {
		query = `
select a.id, 
a.name, 
coalesce(a.phone, '') as phone, 
coalesce(safe_stop_bullying_reports_active, false) as has_incident_reports,
coalesce(safe_stop_lost_item_reports_active, false) as has_lost_item_reports,
coalesce(active, false) as active,
coalesce(a.student_scanning, false) as student_scanning
from jurisdictions a
where a.active = true
order by name;
`
		rows, err := database.GetDB().Queryx(query)
		if err != nil {
			return nil
		} else {
			if rows != nil {
				for rows.Next() {
					j := ClientJurisdiction{}
					err = rows.StructScan(&j)
					if err != nil {
						return nil
					} else {
						client_jurisdictions.Jurisdictions = append(client_jurisdictions.Jurisdictions, j)
					}
				}
			}
		}

		return &client_jurisdictions


	} else if (UserHasAnyPermissionGroups([]string{ pg.License_1, pg.License_2, pg.License_3, pg.License_4}, u.PermissionGroups)){
		query = `
select a.id, 
a.name, 
coalesce(a.phone, '') as phone,
coalesce(safe_stop_bullying_reports_active, false) as has_incident_reports,
coalesce(safe_stop_lost_item_reports_active, false) as has_lost_item_reports,
coalesce(active, false) as active,
coalesce(a.student_scanning, false) as student_scanning
from jurisdictions a 
join jurisdictional_restrictions b on a.id = b.jurisdiction_id
where b.user_id = $1
order by name;
`
	} else {
		query = `
select distinct id, 
name, 
coalesce(phone, '') as phone,
coalesce(safe_stop_bullying_reports_active, false) as has_incident_reports,
coalesce(safe_stop_lost_item_reports_active, false) as has_lost_item_reports,
coalesce(active, false) as active,
coalesce(student_scanning, false) as student_scanning
from (
select a.id, 
a.name, 
a.phone, 
a.safe_stop_lost_item_reports_active, 
a.safe_stop_bullying_reports_active,
a.active,
a.student_scanning
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

select a.id, 
a.name, 
a.phone, 
a.safe_stop_lost_item_reports_active, 
a.safe_stop_bullying_reports_active,
a.active,
a.student_scanning
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
order by z.name
`
	}

	rows, err := database.GetDB().Queryx(query, u.Id)
	if err != nil {
		return nil
	} else {
		if rows != nil {
			for rows.Next() {
				j := ClientJurisdiction{}
				err = rows.StructScan(&j)
				if err != nil {
					return nil
				} else {
					client_jurisdictions.Jurisdictions = append(client_jurisdictions.Jurisdictions, j)
				}
			}
		}
	}

	return &client_jurisdictions
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




func ActivateJurisdiction(id int) interface{} {

	j := struct {
		Id int `json:"id" db:"id"`
		RegistrationText string `json:"registration_text" db:"registration_text"`
		RegistrationImageUrl string `json:"registration_image_url" db:"registration_image_url"`
		RegistrationType string `json:"registration_type" db:"registration_type"`
		RegistrationLabel string `json:"registration_label" db:"registration_label"`
		Ad interface{}
	} {

	}

	query := `
select a.id,
coalesce(a.registration_text, '') as registration_text,
coalesce(a.registration_image_url, '') as registration_image_url,
b.name as registration_type,
a.student_registration_label as registration_label
from jurisdictions a
join safe_stop_registration_types b on a.safe_stop_registration_type_id = b.id
where a.active = true
and a.id = $1;
`
	err := database.GetDB().QueryRowx(query, id).StructScan(&j)
	if err != nil {
		fmt.Print(err)
	}

	j.Ad = NextRegistrationAd(id)

	return j
}

func ActiveProductIdForJurisdiction(jurisdiction_id int) int {
	product_id := 0

	query := `
select a.id
from products a 
join jurisdictions b on a.jurisdiction_id = b.id
join time_zones c on c.id = b.time_zone_id
where a.jurisdiction_id = $1
and a.availability_start_date <= NOW() at time zone c.postgresql_name
and a.availability_end_date > NOW() at time zone c.postgresql_name
and a.active = 't'
and a.product_type = 'ss'
`
	err := database.GetDB().QueryRow(query, jurisdiction_id).Scan(&product_id)
	if err != nil {
		log.Fatal(err)
	}

	return product_id
}


func JurisdictionUsesScanCodeMapping(jurisdiction_id int) bool {
	use_scan_code_mapping := false
	query := `
select coalesce(use_scan_code_mapping, false) as use_scan_code_mapping
from jurisdictions 
where id = $1
`
	err := database.GetDB().QueryRow(query, jurisdiction_id).Scan(&use_scan_code_mapping)
	if err != nil {
		log.Println(err)
	}
	return use_scan_code_mapping

}








