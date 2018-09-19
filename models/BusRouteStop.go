package models

import (
	"github.com/schoolwheels/safestopclient/database"
	"log"
)


type StopLimitInfo struct {
	JurisdictionId int `db:"jurisdiction_id"`
	StopLimit int `db:"stop_limit"`
	GateKey string `db:"gate_key"`
}

func StopLimitInfoForBusRouteStopId(bus_route_stop_id int) *StopLimitInfo{

	r := StopLimitInfo{}

	sql := `
select 
c.id as jurisdiction_id, 
c.selected_stop_limit as stop_limit, 
c.gate_key from
bus_route_stops a
join bus_routes b on a.bus_route_id = b.id
join jurisdictions c on c.id = b.jurisdiction_id
where a.id = $1
limit 1
`
	rows, err := database.GetDB().Queryx(sql, bus_route_stop_id)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	for i := 0; rows.Next(); i++ {
		err = rows.StructScan(&r)
		if err != nil {
			log.Println(err.Error())
			return nil
		}
	}

	return &r
}
