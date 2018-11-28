package models

import "github.com/schoolwheels/safestopclient/database"

type SafeStopInquiry struct {
	Id int
	Email string
	FirstName string
	LastName string
	Phone string
	SchoolOrDistrict string
	PostalCode string
	City string
	State string
	InterestedAsA string
	HowDidYouHearAboutUs string
	CommentsQuestions string
	SchoolOrDistrictEmployee bool
}

func InsertSafeStopInquiry(ssi *SafeStopInquiry) bool {
	query := `
insert into safe_stop_inquiries
(
first_name,
last_name,
email,
school_or_district,
city,
state,
created_at,
updated_at,
school_or_district_employee
) values (
$1,
$2,
$3,
$4,
$5,
$6,
now(),
now(),
$7,
)
`
	_, err := database.GetDB().Exec(
		query,
		ssi.FirstName,
		ssi.LastName,
		ssi.Email,
		ssi.SchoolOrDistrict,
		ssi.City,
		ssi.State,
		ssi.SchoolOrDistrictEmployee,
	)
	if err != nil {
		return false
	}
	return true
}