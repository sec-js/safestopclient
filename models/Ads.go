package models

import (
	"github.com/schoolwheels/safestopclient/database"
	)


func NextRegistrationAd(jurisdiction_id int) interface{} {

	ad := struct {
		Id int `json:"id" db:"id"`
		LoginImageUrl string `json:"login_image_url" db:"login_image"`
	} {

	}

	ads_without_impressions_ct := 0
	query := `
select count(*) as ads_without_impressions_ct
from ads a 
join ads_jurisdictions b 
on b.ad_id = a.id
where b.jurisdiction_id = $1
and a.start_date <= now()::date 
and a.stop_date >= now()::date
and a.login_image is not null
and a.login_image != ''
and a.id not in (select ad_id from ad_impressions)
`
	err := database.GetDB().QueryRow(query, jurisdiction_id).Scan(&ads_without_impressions_ct)
	if err != nil {
		return ad
	}
	if ads_without_impressions_ct > 0 {
		query = `
select a.id,
a.login_image
from ads a 
join ads_jurisdictions b 
on b.ad_id = a.id
where b.jurisdiction_id = $1
and a.start_date <= now()::date 
and a.stop_date >= now()::date
and a.login_image is not null
and a.login_image != ''
and a.id not in (select ad_id from ad_impressions)
limit 1
`
		err := database.GetDB().QueryRowx(query, jurisdiction_id).StructScan(&ad)
		if err != nil {
			return ad
		}
	} else {
		query = `
select a.id,
a.login_image
from ads a 
join ads_jurisdictions b on b.ad_id = a.id
join ad_impressions c on c.ad_id = a.id
where b.jurisdiction_id = $1
and a.start_date <= now()::date 
and a.stop_date >= now()::date
and a.login_image is not null
and a.login_image != ''
group by a.id, a.login_image
order by max(c.id)
limit 1
`
		err := database.GetDB().QueryRowx(query, jurisdiction_id).StructScan(&ad)
		if err != nil {
			return ad
		}
	}
	return ad
}
