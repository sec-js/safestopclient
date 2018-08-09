package models

import (
	"github.com/schoolwheels/safestopclient/database"
	"log"
)

type ScanNotificationSubscriptions struct {
	Subscriptions []ScanNotificationSubscription
}

type ScanNotificationSubscription struct {
	Id int `json:"id" db:"id"`
	UserId int `json:"user_id" db:"user_id"`
	Code string `json:"code" db:"code"`
	Name string `json:"name" db:"name"`
	JurisdictionId int `json:"jurisdiction_id" db:"jurisdiction_id"`
}

func ScanNotificationSubscriptionsForUser(user_id int) *ScanNotificationSubscriptions {

	scan_notification_subscriptions := ScanNotificationSubscriptions{}

	query := `
select a.id,
a.user_id, 
a.code,
a.name, 
a.jurisdiction_id
from bus_rider_scan_notification_subscriptions a 
join jurisdictions b on b.id = a.jurisdiction_id
where b.active = true
and a.user_id = $1
order by a.name;
`
	rows, err := database.GetDB().Queryx(query, user_id)
	if err != nil || rows == nil {
		log.Println(err.Error())
		return nil
	}

	for rows.Next() {
		s := ScanNotificationSubscription{}
		err = rows.StructScan(&s)
		if err != nil {
			log.Println(err.Error())
			return nil
		}
		scan_notification_subscriptions.Subscriptions = append(scan_notification_subscriptions.Subscriptions, s)
	}

	return &scan_notification_subscriptions
}


func InsertScanNotificationSubscriptions(s *ScanNotificationSubscription) bool {

	if s.JurisdictionId == 0 || len(s.Code) == 0 || len(s.Name) == 0 || s.UserId == 0 {
		return false
	}

	use_mapping := JurisdictionUsesScanCodeMapping(s.JurisdictionId)
	if use_mapping {
		new_code := StudentScanCode(s.Code, s.JurisdictionId)
		if new_code == "" {
			return false
		}
		s.Code = new_code
	}

	query := `
insert into bus_rider_scan_notification_subscriptions
(
jurisdiction_id,
user_id,
code,
name,
created_at,
updated_at
) values (
$1,
$2,
$3,
$4,
now(),
now()
)
`
	_, err := database.GetDB().Exec(
		query,
		s.JurisdictionId,
		s.UserId,
		s.Code,
		s.Name)
	if err != nil {
		return false
	}
	return true
}

func DeleteScanNotificationSubscriptions(id int) bool {
	if id == 0 {
		return false
	}

	query := `delete from bus_rider_scan_notification_subscriptions where id = $1`
	_, err := database.GetDB().Exec(query, id)
	if err != nil {
		return false
	}

	return true
}