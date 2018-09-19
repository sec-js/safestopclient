package models

import "github.com/schoolwheels/safestopclient/database"


type ScanNotification struct {
	Id int `json:"id"`
	DateOccurred string `json:"date_occurred"`
	Name string `json:"name"`
}

func DismissScanNotification(scan_notification_id int) bool {
	if scan_notification_id > 0 {
		query := `update bus_rider_scan_notifications set dismissed_at = now() where id = $1`
		_, err := database.GetDB().Exec(query, scan_notification_id)
		if err != nil {
			return false
		}
		return true
	}
	return false
}
