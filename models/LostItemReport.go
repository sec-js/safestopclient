package models

import "github.com/schoolwheels/safestopclient/database"

type LostItemReport struct {
	Id int
	Email string
	JurisdictionId int
	JurisdictionName string
	Description string
	DateLost string
	RouteIdentifier string
	LastName string
	FirstName string
	Phone string
	Summary string
}


func InsertLostItemReport(lir *LostItemReport) bool {
	query := `
insert into safe_stop_lost_item_reports
(
jurisdiction_id,
first_name,
last_name,
email_address,
phone_number,
description,
date_lost,
route_identifier,
--summary,
status,
created_at,
updated_at
) values (
$1,
$2,
$3,
$4,
$5,
$6,
$7,
$8,
'open',
now(),
now()
)
`
	_, err := database.GetDB().Exec(
		query,
		lir.JurisdictionId,
		lir.FirstName,
		lir.LastName,
		lir.Email,
		lir.Phone,
		lir.Description,
		lir.DateLost,
		lir.RouteIdentifier,
		//lir.Summary
		)
	if err != nil {
		return false
	}
	return true
}