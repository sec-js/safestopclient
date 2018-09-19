package controllers

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/schoolwheels/safestopclient/models"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"strconv"
)

type APIController struct {
	*ControllerBase
}

func (c *APIController) Register() {

	c.RoutePrefix = "/api"

	//actions
	c.addRouteWithPrefix("/version", c.versionAction)
	c.addRouteWithPrefix("/student_exists", c.StudentExistsAction)
	c.addRouteWithPrefix("/school_code_exists", c.SchoolCodeExistsAction)
	c.addRouteWithPrefix("/email_exists", c.EmailExistsAction)
	c.addRouteWithPrefix("/test", c.TestAction)
	c.addRouteWithPrefix("/test_google", c.TestGoogleAction)
	c.addRouteWithPrefix("/available_bus_routes", c.AvailableBusRoutesAction)
	c.addRouteWithPrefix("/available_bus_route_stops", c.AvailableBusRouteStopsAction)
	c.addRouteWithPrefix("/add_user_stop", c.AddUserStopAction)
	c.addRouteWithPrefix("/remove_user_stop", c.RemoveUserStopAction)
	c.addRouteWithPrefix("/alerts", c.AlertsAction)
	c.addRouteWithPrefix("/set_viewed_alerts", c.SetViewedAlertsAction)
	c.addRouteWithPrefix("/my_stops", c.MyStopsAction)
	c.addRouteWithPrefix("/map", c.MapAction)
	c.addRouteWithPrefix("/scan_notifications", c.ScanNotificationsAction)
	c.addRouteWithPrefix("/dismiss_scan_notification", c.DismissScanNotificationAction)

	c.addRouteWithPrefix("/next_ad", c.NextAdAction)




}

func (c *APIController) versionAction(w http.ResponseWriter, r *http.Request) {

	data := struct {
		Version float64 `json:"version"`
	}{
		1.0,
	}

	c.renderJSON(data, w)
}















//http://ssc.local:8080/api/student_exists?sis_identifier=112408&jurisdiction_id%20=%2015
func (c *APIController) StudentExistsAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.ParseForm()

	sis_identifier := r.FormValue("sis_identifier")
	jurisdiction_id, err :=  strconv.Atoi(r.FormValue("jurisdiction_id"))
	if err != nil || sis_identifier == "" {
		v := struct {
			Valid bool `json:"valid"`
		} {
			false,
		}
		c.renderJSON(v, w)
	} else {
		v := struct {
			Valid bool `json:"valid"`
		} {
			models.StudentIdentifierExists(sis_identifier, jurisdiction_id),
		}

		c.renderJSON(v, w)
	}
}

//http://ssc.local:8080/api/school_code_exists?school_code=MACC48&jurisdiction_id=214
func (c *APIController) SchoolCodeExistsAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	r.ParseForm()

	school_code := r.FormValue("school_code")
	jurisdiction_id, err :=  strconv.Atoi(r.FormValue("jurisdiction_id"))
	if err != nil || school_code == "" {
		v := struct {
			Valid bool `json:"valid"`
		} {
			false,
		}
		c.renderJSON(v, w)
	} else {
		v := struct {
			Valid bool `json:"valid"`
		} {
			models.SchoolCodeExists(school_code, jurisdiction_id),
		}
		c.renderJSON(v, w)
	}
}

//http://ssc.local:8080/api/email_exists?user[email]=acook@ridesta.comfff
func (c *APIController) EmailExistsAction(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	v := struct {
		Valid bool `json:"valid"`
	} {
		!models.EmailExists(r.FormValue("user[email]")),
	}
	w.Header().Set("Content-Type", "application/json")
	c.renderJSON(v, w)
}






func (c *APIController) AvailableBusRoutesAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}
	u := models.FindUser(uid)


	page, err := strconv.Atoi(r.FormValue("page"))
	if err != nil {
		page = 1
	}

	if u.SuperAdmin == true || models.UserHasAnyPermissionGroups([]string{
		c.PermissionGroups.Admin,
		c.PermissionGroups.License_1,
		c.PermissionGroups.License_2,
		c.PermissionGroups.License_3,
		c.PermissionGroups.License_4,
	}, u.PermissionGroups) {
		c.renderJSON(models.BusRoutesForAdminAndSchoolAdminUser(page, r.FormValue("search"), r.FormValue("address_1"), r.FormValue("postal_code"), u, c.PermissionGroups), w)
	} else {
		c.renderJSON(models.BusRoutesForRegularUsers(page, r.FormValue("address_1"), r.FormValue("postal_code"), u, c.PermissionGroups), w)
	}
}




func (c *APIController) AvailableBusRouteStopsAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}
	u := models.FindUser(uid)

	bus_route_id, err := strconv.Atoi(r.FormValue("bus_route_id"))
	if err != nil {
		bus_route_id = -1
	}

	sl := []models.UserJurisdictionStopLimit{}
	if u.SuperAdmin == false && models.UserHasAnyPermissionGroups(
		[]string{
			c.PermissionGroups.Admin,
			c.PermissionGroups.License_1,
			c.PermissionGroups.License_2,
			c.PermissionGroups.License_3,
			c.PermissionGroups.License_4,
			}, u.PermissionGroups) == false {
		sl = *models.UsersStopLimits(uid)
	}

	c.renderJSON(
		struct {
			UserStopLimitations []models.UserJurisdictionStopLimit `json:"user_stop_limitations"`
			BusRouteStops []models.BusRouteStopSearchResult `json:"bus_route_stops"`
		}{
			sl,
			models.BusRouteStopsForUser(uid, bus_route_id),
		}, w)

	return;
}












func (c *APIController) AddUserStopAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}
	u := models.FindUser(uid)

	bus_route_stop_id, err := strconv.Atoi(r.FormValue("bus_route_stop_id"))
	if err != nil {

	}

	if u.SuperAdmin == false && models.UserHasAnyPermissionGroups(
		[]string{
			c.PermissionGroups.Admin,
			c.PermissionGroups.License_1,
			c.PermissionGroups.License_2,
			c.PermissionGroups.License_3,
			c.PermissionGroups.License_4,
		}, u.PermissionGroups) == false {
		models.AddStopToRegularUsers(bus_route_stop_id, u.Id)
		c.renderJSON(*models.UsersStopLimits(uid), w)
		return;
	} else {
		models.AddStopToAdminAndSchoolUsers(bus_route_stop_id, u.Id)
		c.renderJSON([]models.UserJurisdictionStopLimit{}, w)
		return;
	}
}


func (c *APIController) RemoveUserStopAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}
	u := models.FindUser(uid)

	bus_route_stop_id, err := strconv.Atoi(r.FormValue("bus_route_stop_id"))
	if err != nil {

	}

	models.DeleteBusRouteStopUser(bus_route_stop_id, u.Id)

	if u.SuperAdmin == false && models.UserHasAnyPermissionGroups(
		[]string{
			c.PermissionGroups.Admin,
			c.PermissionGroups.License_1,
			c.PermissionGroups.License_2,
			c.PermissionGroups.License_3,
			c.PermissionGroups.License_4,
		}, u.PermissionGroups) == false {
		c.renderJSON(*models.UsersStopLimits(uid), w)
		return;
	} else {
		c.renderJSON([]models.UserJurisdictionStopLimit{}, w)
		return;
	}
}



















//http://ssc.local:8080/api/next_registration_ad?jurisdiction_id=172
func (c *APIController) TestAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	aws_config := aws.Config{
		Credentials: credentials.NewEnvCredentials(),
		Region: aws.String(os.Getenv("AWS_REGION")),
	}


	resp := ""
	sess, err := session.NewSession(&aws_config)
	if err != nil {
		resp = err.Error();
		log.Println(err)

	}

	if sess != nil {
		sns_client := sns.New(sess, &aws_config)
		if sns_client != nil {


			end_point, err := sns_client.CreatePlatformEndpoint(&sns.CreatePlatformEndpointInput{
				PlatformApplicationArn: aws.String(viper.GetString("SNS_ANDROID_ARN")),
				Token: aws.String(""),
			})

			if err != nil {
				resp = err.Error()
				log.Println(err)
			}

			if end_point != nil {
				resp = aws.StringValue(end_point.EndpointArn)
			}

		}
	}


	c.renderJSON(resp, w)
}



func (c *APIController) AlertsAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}
	u := models.FindUser(uid)

	viewed_alert_ids := models.ViewedAlertIdsForUserId(u.Id)
	alerts := []models.Alert{}
	unread_messages := 0


	if u.SuperAdmin == true || models.UserHasAnyPermissionGroups([]string{c.PermissionGroups.Admin}, u.PermissionGroups) {
		alerts = models.AlertsForAdminUser()
	} else {
		alerts = models.AlertsForSchoolAdminAndRegularUsers(u, c.PermissionGroups)
	}

	for x:= 0; x < len(alerts); x++ {
		if len(viewed_alert_ids) > 0 {
			for  y := 0; y < len(viewed_alert_ids); y++ {
				if alerts[x].Id == viewed_alert_ids[y] {
					break
				}
				if y == len(viewed_alert_ids) - 1 {
					unread_messages++
				}
			}
		} else {
			unread_messages++
		}
		if unread_messages > 0 {
			break
		}
	}

	c.renderJSON(
		struct {
			Alerts []models.Alert `json:"alerts"`
			UnreadMessages int `json:"unread_messages"`
		}{
			alerts,
			unread_messages,
		}, w)
}


func (c *APIController) SetViewedAlertsAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}
	u := models.FindUser(uid)

	models.SetUsersViewedAlerts(u, r.FormValue("alert_ids"))

	c.renderJSON(struct {}{}, w)
}


func (c *APIController) MyStopsAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}
	u := models.FindUser(uid)

	ms := []models.MyStopsJurisdiction{}
	dbr := models.UsersMyStops(u, c.PermissionGroups)

	if len(dbr) > 0 {
		current_jurisdiction := models.MyStopsJurisdiction{}
		current_route := models.MyStopsRoute{}
		current_stop := models.MyStopsStop{}

		for x := 0; x < len(dbr); x++ {

			if current_jurisdiction.Name == ""{
				current_jurisdiction.Name = dbr[x].JurisdictionName
			} else {
				if current_jurisdiction.Name != dbr[x].JurisdictionName {
					current_route.Stops = append(current_route.Stops, current_stop)
					current_jurisdiction.Routes = append(current_jurisdiction.Routes, current_route)
					ms = append(ms, current_jurisdiction)
					current_route = models.MyStopsRoute{}
					current_jurisdiction = models.MyStopsJurisdiction{ Name: dbr[x].JurisdictionName }
				}
			}


			if current_route.Id != dbr[x].BusRouteId {
				if current_route.Id > 0 {
					current_jurisdiction.Routes = append(current_jurisdiction.Routes, current_route)
				}
				current_route = models.MyStopsRoute{
					Id: dbr[x].BusRouteId,
					Name: dbr[x].BusRouteName,
					Active: dbr[x].BusRouteActive,
					HidePredictions: dbr[x].HidePredictions,
					Audible: dbr[x].Audible,
				}

				if dbr[x].BusRouteActive == false {
					current_route.Errors = true
					current_route.BusAssigned = true
					current_route.Active = false
				} else if dbr[x].BusAssigned == -1 {
					current_route.Errors = true
					current_route.BusAssigned = false
					current_route.Active = true
				} else {
					current_route.Errors = false
					current_route.BusAssigned = true
					current_route.Active = true
					current_route.Shuttle = dbr[x].LoopMode != "off"
				}

				if u.SuperAdmin == true || models.UserHasAnyPermissionGroups([]string{c.PermissionGroups.Admin}, u.PermissionGroups) {
					current_route.HidePredictions = false;
				}
			}

			if current_stop.Id != dbr[x].StopId {
				if current_stop.Id > 0 {
					current_route.Stops = append(current_route.Stops, current_stop)
				}

				current_stop = models.MyStopsStop{
					Id: dbr[x].StopId,
					Name: dbr[x].StopName,
					ScheduledTime: dbr[x].ScheduledTime,
					Latitude: dbr[x].StopLatitude,
					Longitude: dbr[x].StopLongitude,
				}

				if current_route.Errors == false {
					if(len(dbr[x].ArrivalTime) > 0) {
						current_stop.Time = dbr[x].ArrivalTime
						current_stop.TimeClass = "arrived"
						current_stop.TimeTitle = string(T(currentLocale(c.ControllerBase, r),  "arrived", ""))
					} else if len(dbr[x].SkippedAt) > 0 {
						current_stop.TimeTitle = "No Stop Recorded"
						current_stop.AsOf = "No Stop Recorded"
					} else {

						current_stop.TimeTitle = string(T(currentLocale(c.ControllerBase, r),  "expected", ""))
						if dbr[x].PredictedTimeOffset != -1 {
							if dbr[x].HidePredictions == false {
								pto := dbr[x].PredictedTimeOffset
								sto := dbr[x].ScheduledTimeOffset
								if dbr[x].PredictedTimeOffset > dbr[x].ScheduledTimeOffset {
									dif := (pto - sto) / 60
									if dif >= 0 && dif <= 5 {
										current_stop.TimeClass = "expected-on-time"
									} else if dif >= 6 && dif <= 15 {
										current_stop.TimeClass = "expected-late"
									} else if dif > 15 {
										current_stop.TimeClass = "expected-really-late"
									}
								}
							}
							current_stop.Time = dbr[x].PredictedTimeString
							current_stop.AsOf = ""
						} else {
							current_stop.Time = dbr[x].ScheduledTime
							current_stop.TimeClass = "expected-on-time"
							current_stop.AsOf = ""
						}
					}
				}
			}
		}

		current_route.Stops = append(current_route.Stops, current_stop)
		current_jurisdiction.Routes = append(current_jurisdiction.Routes, current_route)
		ms = append(ms, current_jurisdiction)
	}


	c.renderJSON(ms, w)
}






func (c *APIController) MapAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stops := []models.MapViewStop{}

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}
	u := models.FindUser(uid)

	dbr := models.UsersMyStops(u, c.PermissionGroups)

	if len(dbr) > 0 {

		for x := 0; x < len(dbr); x++ {
			if r.FormValue("bus_route_id") == strconv.Itoa(dbr[x].BusRouteId) {
				s := models.MapViewStop{}
				s.StopId = dbr[x].StopId
				s.StopName = dbr[x].StopName
				s.StopLatitude = dbr[x].StopLatitude
				s.StopLongitude = dbr[x].StopLongitude
				s.StopScheduledTime = dbr[x].ScheduledTime
				s.BusRouteId = dbr[x].BusRouteId
				s.BusRouteName = dbr[x].BusRouteName
				s.BusLatitude = dbr[x].BusLatitude
				s.BusLongitude = dbr[x].BusLongitude
				s.Audible = dbr[x].Audible


				if u.SuperAdmin == true || models.UserHasAnyPermissionGroups([]string{c.PermissionGroups.Admin}, u.PermissionGroups) {
					s.HidePredictions = false
					s.ShowBus = true
				}

				s.Shuttle = dbr[x].LoopMode != "off"

				if (len(dbr[x].ArrivalTime) > 0) {
					s.Time = dbr[x].ArrivalTime
					s.TimeClass = "arrived"
					s.TimeTitle = string(T(currentLocale(c.ControllerBase, r), "arrived", ""))
				} else if len(dbr[x].SkippedAt) > 0 {
					s.TimeTitle = "No Stop Recorded"
					s.AsOf = "No Stop Recorded"
				} else {

					s.TimeTitle = string(T(currentLocale(c.ControllerBase, r), "expected", ""))
					if dbr[x].PredictedTimeOffset != -1 {
						if dbr[x].HidePredictions == false {
							pto := dbr[x].PredictedTimeOffset
							sto := dbr[x].ScheduledTimeOffset
							if dbr[x].PredictedTimeOffset > dbr[x].ScheduledTimeOffset {
								dif := (pto - sto) / 60
								if dif >= 0 && dif <= 5 {
									s.TimeClass = "expected-on-time"
								} else if dif >= 6 && dif <= 15 {
									s.TimeClass = "expected-late"
								} else if dif > 15 {
									s.TimeClass = "expected-really-late"
								}
							}
						}
						s.Time = dbr[x].PredictedTimeString
						s.AsOf = ""
					} else {
						s.Time = dbr[x].ScheduledTime
						s.TimeClass = "expected-on-time"
						s.AsOf = ""
					}
				}
				stops = append(stops, s)
			}
		}
	}
	c.renderJSON(stops, w)
}






func (c *APIController) ScanNotificationsAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}
	u := models.FindUser(uid)
	c.renderJSON(models.UserScanNotifications(u), w)
}


func (c *APIController) DismissScanNotificationAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}
	u := models.FindUser(uid)

	scan_id, err := strconv.Atoi(r.FormValue("scan_notification_id"))
	if err != nil {

	}

	if u.Id > 0 {
		c.renderJSON(models.DismissScanNotification(scan_id), w)
	} else {
		c.renderJSON(false, w)
	}
}


func (c *APIController) NextAdAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}
	u := models.FindUser(uid)
	c.renderJSON(models.NextAd(u, c.PermissionGroups), w)
}


func (c *APIController) GCMAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	resp := struct {
		Success bool `json:"success"`
	} {
		false,
	}

	device_platform := "Android"
	device_token := r.FormValue("device_token")

	if device_token != "" {

		if models.InsertDevice(device_platform, device_token) == false {
			c.renderJSON(resp, w)
			return
		}


		end_point_arn := ""

		aws_config := aws.Config{
			Credentials: credentials.NewEnvCredentials(),
			Region:      aws.String(viper.GetString("SNS_REGION")),
		}

		sess, err := session.NewSession(&aws_config)
		if err != nil {
			c.renderJSON(resp, w)
			return
		}

		if sess == nil {
			c.renderJSON(resp, w)
			return
		}

		if viper.GetString("domain") == "safestopapp.ca" {

			lambda_client := lambda.New(sess)
			if lambda_client == nil {
				c.renderJSON(resp, w)
				return
			}



		} else if (viper.GetString("domain") == "safestopapp.com") {

			sns_client := sns.New(sess, &aws_config)
			if sns_client == nil {
				c.renderJSON(resp, w)
				return
			}

			end_point, err := sns_client.CreatePlatformEndpoint(&sns.CreatePlatformEndpointInput{
				PlatformApplicationArn: aws.String(viper.GetString("SNS_ANDROID_ARN")),
				Token: aws.String(device_token),
			})

			if err != nil {
				c.renderJSON(resp, w)
				return
			}

			end_point_arn = aws.StringValue(end_point.EndpointArn)

		} else {

		}


		resp.Success = models.UpdateDeviceARN(device_platform, device_token, aws.StringValue(end_point.EndpointArn))

	}

	c.renderJSON(resp, w)
}


func (c *APIController) GCMUserAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c.renderJSON(models.Geocode("1680 Eider Down Dr.", "29483"), w)
}






func (c *APIController) TestGoogleAction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	c.renderJSON(models.Geocode("1680 Eider Down Dr.", "29483"), w)
}
