package models

import "github.com/schoolwheels/safestopclient/database"

type ServiceIssue struct {
	UserId int
	Email string
	JurisdictionId int
	JurisdictionName string
	IssueType string
	Description string
	Date string
}

func InsertServiceIssue(si *ServiceIssue) bool {
	query := `
insert into safe_stop_service_issues
(
issue_type,
description,
user_id,
jurisdiction_id,
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
		si.IssueType,
		si.Description,
		si.UserId,
		si.JurisdictionId)
	if err != nil {
		return false
	}
	return true
}
