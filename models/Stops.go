package models

type MyStop struct {

	Id int `json:"id" db:"id"`
	Name string `json:"stop_name" db:"stop_name"`
	Latitude float64 `json:"latitude" db:"latitude"`
	Longitude float64 `json:"longitude" db:"longitude"`
	ScheduledTime string `json:"scheduled_time" db:"scheduled_time"`
	RouteId string `json:"route_id"`
	RouteName string `json:"route_name" db:"route_name"`


	BusId int `json:"bus_id" db:"bus_id"`
	BusName string `json:"bus_name" db:"bus_name"`
	BusLatitude float64 `json:"bus_latitude" db:"bus_latitude"`
	BusLongitude float64 `json:"bus_longitude" db:"bus_longitude"`

	JurisdictionId int `json:"jurisdiction_id" db:"jurisdiction_id"`
	JurisdictionName int `json:"jurisdiction_name" db:jurisdiction_name`





}
