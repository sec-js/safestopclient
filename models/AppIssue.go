package models

import "github.com/schoolwheels/safestopclient/database"

type AppIssue struct {
	UserId int
	Email string
	JurisdictionId int
	JurisdictionName string
	IssueType string
	Description string
	Date string
}

func InsertAppIssue(ai *AppIssue) bool {
	query := `
insert into safe_stop_app_issues
(
jurisdiction_id,
user_id,
issue_type,
description,
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
		ai.JurisdictionId,
		ai.UserId,
		ai.IssueType,
		ai.Description)
	if err != nil {
		return false
	}
	return true
}
