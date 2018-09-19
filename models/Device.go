package models

import "github.com/schoolwheels/safestopclient/database"

func InsertDevice(device_platform string, device_token string) bool {

	if device_platform == "" || device_token == "" {
		return false
	}

	query := `
insert into devices
(
device_platform,
notification_token,
created_at,
updated_at
) values (
$1,
$2,
now(),
now()
)
`
	_, err := database.GetDB().Exec(query, device_platform, device_token)
	if err != nil {
		return false
	}
	return true
}


func UpdateDeviceARN(device_platform string, device_token string, arn string) bool {

	if device_platform == "" || device_token == "" || arn == "" {
		return false
	}

	query := `update devices set sns_arn = $1 where notification_token = $2 and device_platform = $3`

	_, err := database.GetDB().Exec(query, arn, device_token, device_platform)
	if err != nil {
		return false
	}
	return true
}

func UpdateDeviceARNAndUser(device_platform string, device_token string, arn string, user_id int) bool {

	if device_platform == "" || device_token == "" || arn == "" || user_id <= 0 {
		return false
	}

	query := `update devices set sns_arn = $1, user_id = $2 where notification_token = $3 and device_platform = $4`

	_, err := database.GetDB().Exec(query, arn, user_id, device_token, device_platform)
	if err != nil {
		return false
	}
	return true
}
