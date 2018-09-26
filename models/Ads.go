package models

import (
	"github.com/schoolwheels/safestopclient/database"
	)

type Ad struct {
	Id int `db:"id" json:"id"`
	Url string `db:"url" json:"url"`
	TargetUrl string `db:"target_url" json:"target_url"`
}

func NextAd(u *User, pg *PermissionGroups) Ad {

	jurisdiction_ids := UsersClientJurisdictionIds(u, pg)

	ad := Ad{}

	ads_without_impressions_ct := 0
	query := `
select count(*) as ads_without_impressions_ct
from ads a 
join ads_jurisdictions b 
on b.ad_id = a.id
where b.jurisdiction_id in (` + jurisdiction_ids + `)
and a.start_date <= now()::date 
and a.stop_date >= now()::date
and a.app_image is not null
and a.app_image != ''
and a.id not in (select coalesce(ad_id, -1) from ad_impressions)
`
	err := database.GetDB().QueryRow(query).Scan(&ads_without_impressions_ct)
	if err != nil {
		return ad
	}

	if ads_without_impressions_ct > 0 {
		query = `
select a.id,
a.app_image as url,
a.target_url as target_url
from ads a 
join ads_jurisdictions b 
on b.ad_id = a.id
where b.jurisdiction_id in (` + jurisdiction_ids + `)
and a.start_date <= now()::date 
and a.stop_date >= now()::date
and a.app_image is not null
and a.app_image != ''
and a.id not in (select coalesce(ad_id, -1) from ad_impressions)
limit 1
`

		err := database.GetDB().QueryRowx(query).StructScan(&ad)
		if err != nil {
			return ad
		}
	} else {
		query = `
select a.id,
a.app_image as url,
a.target_url as target_url
from ads a 
join ads_jurisdictions b on b.ad_id = a.id
join ad_impressions c on c.ad_id = a.id
where b.jurisdiction_id in (` + jurisdiction_ids + `)
and a.start_date <= now()::date 
and a.stop_date >= now()::date
and a.app_image is not null
and a.app_image != ''
group by a.id, a.login_image
order by max(c.id)
limit 1
`
		err := database.GetDB().QueryRowx(query).StructScan(&ad)
		if err != nil {
			return ad
		}
	}
	return ad
}





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
