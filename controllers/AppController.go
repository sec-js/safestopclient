package controllers

import (
	"net/http"
	"github.com/schoolwheels/safestopclient/models"
	"strings"
	"reflect"
	"github.com/gorilla/mux"
	"strconv"
	"github.com/schoolwheels/safestopclient/database"
)

type AppController struct {
	*ControllerBase
}

func (c *AppController) Register() {

	//templates
	c.addTemplate("index", "index.html", "app.html")
	c.addTemplate("check_availability", "check_availability.html", "default.html")
	c.addTemplate("account", "account.html", "app.html")
	c.addTemplate( "language", "language.html", "app.html")
	c.addTemplate( "activate", "activate.html", "default.html")
	c.addTemplate("faq", "faq.html", "default.html")
	c.addTemplate("failed_registration_attempt", "failed_registration_attempt.html", "default.html")
	c.addTemplate("report_an_app_issue", "report_an_app_issue.html", "default.html")
	c.addTemplate( "manage_notifications", "manage_notifications.html", "default.html")
	c.addTemplate( "manage_subscriptions", "manage_subscriptions.html", "default.html")
	c.addTemplate( "subscription_details", "subscription_details.html", "default.html")


	c.addTemplate("lost_item_report", "lost_item_report.html", "default.html")

	//actions
	c.addRouteWithPrefix("/", c.IndexAction)
	c.addRouteWithPrefix("/check_availability", c.CheckAvailabilityAction)
	c.addRouteWithPrefix("/change_locale/{locale}", c.ChangeLocaleAction)
	c.addRouteWithPrefix("/account", c.AccountAction)
	c.addRouteWithPrefix( "/language", c.LanguageAction)
	c.addRouteWithPrefix("/activate/{jurisdiction_id}", c.ActivateAction)
	c.addRouteWithPrefix("/faq", c.FaqAction)
	c.addRouteWithPrefix("/failed_registration_attempt", c.FailedRegistrationAttemptAction)
	c.addRouteWithPrefix("/report_an_app_issue", c.AppIssueAction)
	c.addRouteWithPrefix( "/remove_all_stops", c.RemoveAllStopsAction)
	c.addRouteWithPrefix( "/manage_notifications", c.ManageNotificationsAction)
	c.addRouteWithPrefix( "/manage_subscriptions", c.ManageSubscriptionsAction)
	c.addRouteWithPrefix( "/subscription_details/{subscription_id}", c.SubscriptionDetailsAction)

	c.addRouteWithPrefix( "/add_scan_notification_subscription", c.AddScanNotificationSubscriptionAction)
	c.addRouteWithPrefix( "/remove_scan_notification_subscription", c.RemoveScanNotificationSubscriptionAction)

	c.addRouteWithPrefix( "/add_student", c.AddStudentAction)
	c.addRouteWithPrefix("/remove_student", c.RemoveStudentAction)

	c.addRouteWithPrefix("/add_sub_account_user", c.AddSubAccountUserAction)


	c.addRouteWithPrefix("/lost_item_report", c.LostItemReportAction)

}

type dashData struct {
	CurrentUserEmail string
}

func (c *AppController) IndexAction(w http.ResponseWriter, r *http.Request) {

	data := struct {
		Token string
		Email string
	} {
		r.FormValue("token"),
		r.FormValue("email"),
	}


	session, _ :=  c.SessionStore.Get(r, "auth")
	email := session.Values["current_user_email"]
	if email != nil {
		http.Redirect(w, r, r.URL.Host+"/dashboard", http.StatusFound)
		return
	}
	c.render(w, r, "index", data)
}

func (c *AppController) AccountAction(w http.ResponseWriter, r *http.Request) {

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	u := models.FindUser(uid)
	cj := models.ClientJurisdictionForUser(u, c.ControllerBase.PermissionGroups)

	has_jurisdictions := false
	if cj != nil && len(cj.Jurisdictions) > 0 {
		has_jurisdictions = true
	}

	view_manage_notifications := models.UserHasAnyPermissionGroups([]string{
		c.PermissionGroups.Admin,
		c.PermissionGroups.License_3,
		c.PermissionGroups.License_4,
	}, u.PermissionGroups)

	if u.SuperAdmin == true {
		view_manage_notifications = true
	}

	view_manage_subscriptions := models.UserHasAnyPermissionGroups([]string{
		c.PermissionGroups.License_5,
	}, u.PermissionGroups)

	view_lost_item_reports := false
	for i := 0; i < len(cj.Jurisdictions); i++ {
		if cj.Jurisdictions[i].HasLostItemReports == true {
			view_lost_item_reports = true
			break
		}
	}

	data := struct {
		JurisdictionCount int
		HasJurisdictions bool
		Jurisdictions *models.ClientJurisdictions
		ViewManageNotifications bool
		ViewManageSubscriptions bool
		ViewReportLostItem bool
	}{
		len(cj.Jurisdictions),
		has_jurisdictions,
		cj,
		view_manage_notifications,
		view_manage_subscriptions,
		view_lost_item_reports,
	}

	c.render(w, r, "account", data)
}



func (c *AppController) CheckAvailabilityAction(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {

		data := struct {
			PostalCode string
			Jurisdictions *models.Jurisdictions
			JurisdictionCount int
		} {
			r.FormValue("postal_code"),
			nil,
			-1,
		}

		c.render(w, r, "check_availability", data)

	} else {

		postal_code := r.FormValue("postal_code")
		jurisdictions := &models.Jurisdictions{}

		pcr := models.PostalCodeReferenceForPostalCode(postal_code)
		if pcr != nil {
			s := models.StateForAbbreviation(pcr.StateCode)
			if s != nil {
				jurisdictions = models.AvailableJurisdictionsForState(s.Id, postal_code)
			}
		}

		data := struct {
			PostalCode string
			Jurisdictions *models.Jurisdictions
			JurisdictionCount int
		} {
			r.FormValue("postal_code"),
			jurisdictions,
			len(jurisdictions.Jurisdictions),
		}

		c.render(w, r, "check_availability", data)

	}

}









func (c *AppController) ActivateAction(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	postal_code := r.FormValue("postal_code")

	user_id := currentUserId(c.ControllerBase, r)
	if user_id == 0 {
		http.Redirect(w, r, r.URL.Host+"/check_availability?postal_code=" + postal_code, http.StatusFound)
		return
	}

	user := models.FindUser(user_id)

	jurisdiction_id, err := strconv.Atoi(vars["jurisdiction_id"])
	if err != nil {
		setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "jurisdiction_not_available", "")), c.BootstrapAlertClass.Danger)
		http.Redirect(w, r, r.URL.Host+"/check_availability?postal_code=" + postal_code, http.StatusFound)
		return
	}

	if models.UserHasSubscriptionForJurisdiction(user, c.PermissionGroups, jurisdiction_id) {
		setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "already_have_subscription", "")), c.BootstrapAlertClass.Info)
		http.Redirect(w, r, r.URL.Host+"/check_availability?postal_code=" + postal_code, http.StatusFound)
		return
	}



	if r.Method == "GET" {

		data := struct {
			Jurisdiction interface{}
			PostalCode   string
		}{
			models.ActivateJurisdiction(jurisdiction_id),
			r.FormValue("postal_code"),
		}
		c.render(w, r, "activate", data)
		return

	} else {

		r.ParseForm()


		registration_type := r.FormValue("registration_type")

		product_id := models.ActiveProductIdForJurisdiction(jurisdiction_id)
		if product_id == 0 {
			setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "jurisdiction_has_no_products", "")), c.BootstrapAlertClass.Danger)
			http.Redirect(w, r, r.URL.Host+"/check_availability?postal_code=" + postal_code, http.StatusFound)
			return
		}

		if registration_type == "Student Identifier" || registration_type == "Access Code + Student Identifier" {

			student_identifiers := r.Form["student_information[][sis_identifier]"]
			subscription_created, err := models.ActivateStudentIdentifierSubscription(jurisdiction_id, product_id, user, student_identifiers)
			if !subscription_created || err != nil {
				setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "error_while_processing_request", "")), c.BootstrapAlertClass.Danger)
				http.Redirect(w, r, r.URL.Host+"/activate/" + strconv.Itoa(jurisdiction_id) + "?postal_code=" + postal_code, http.StatusFound)
				return
			}

		} else {

			subscription_created, err := models.ActivateAccessCodeSubscription(jurisdiction_id, product_id, user)

			if !subscription_created || err != nil {
				setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "error_while_processing_request", "")), c.BootstrapAlertClass.Danger)
				http.Redirect(w, r, r.URL.Host+"/activate/" + strconv.Itoa(jurisdiction_id) + "?postal_code=" + postal_code, http.StatusFound)
				return
			}

		}

		setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "you_have_successfully_registered_for", "")), c.BootstrapAlertClass.Info)
		http.Redirect(w, r, r.URL.Host+ "/", http.StatusFound)
		return

	}
}

func (c *AppController) ChangeLocaleAction(w http.ResponseWriter, r *http.Request){
	vars := mux.Vars(r)
	session, err:= c.SessionStore.Get(r, "auth")
	session.Values["locale"] = vars["locale"]
	err = session.Save(r, w)
	if err != nil {
		//http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		//return
	}
	http.Redirect(w, r,"/language", http.StatusFound)
}


func (c *AppController) LanguageAction(w http.ResponseWriter, r *http.Request) {

	uid := currentUserId(c.ControllerBase, r)
	if(uid == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}


	c.render(w, r, "language", nil)
}


func (c *AppController) MapAction(w http.ResponseWriter, r *http.Request) {
	session, _ :=  c.SessionStore.Get(r, "auth")
	email := session.Values["current_user_email"]
	if email != nil {
		http.Redirect(w, r, r.URL.Host+"/dashboard", http.StatusFound)
	}

	c.render(w, r, "map", nil)
}

func (c *AppController) FaqAction(w http.ResponseWriter, r *http.Request) {

	type FAQ struct{
		Question string `db:"question"`
		Answer string `db:"answer"`
	}

	data := struct{
		Faq []FAQ
	} {

	}

	query := `select question, answer from frequently_asked_questions`
	rows, err := database.GetDB().Queryx(query)
	if err != nil {

	} else {
		for rows.Next() {
			d := FAQ{}
			rows.StructScan(&d)
			data.Faq = append(data.Faq, d)
		}
	}

	c.render(w, r, "faq", data)
}



func (c *AppController) FailedRegistrationAttemptAction(w http.ResponseWriter, r *http.Request) {

	user_id := currentUserId(c.ControllerBase, r)
	jurisdiction_id := r.FormValue("jurisdiction_id")
	postal_code := r.FormValue("postal_code")

	if user_id == 0 {
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	} else {
		u := models.FindUser(user_id)
		if r.Method == "GET" {
			data := models.FailedRegistrationAttempt{
				JurisdictionId: jurisdiction_id,
				FirstName: u.FirstName,
				LastName: u.LastName,
				Email: u.Email,
				IdOrCodeAttempted: r.FormValue("id_or_code"),
				PostalCode: postal_code,
			}
			c.render(w, r, "failed_registration_attempt", data)
			return
		} else {
			if jurisdiction_id != "" {
				data := models.FailedRegistrationAttempt{
					JurisdictionId: jurisdiction_id,
					FirstName: r.FormValue("first_name"),
					LastName: r.FormValue("last_name"),
					Email: r.FormValue("email"),
					StudentFirstName: r.FormValue("student_first_name"),
					StudentLastName: r.FormValue("student_last_name"),
					IdOrCodeAttempted: r.FormValue("id_or_code_attempted"),
				}
				success := models.InsertFailedRegistrationAttempt(&data)
				if success == true {

					//TODO SEND FAILED REGISTRATION ATTEMPT EMAIL

					setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "request_has_been_submitted", "")), c.BootstrapAlertClass.Info)
					http.Redirect(w, r, r.URL.Host+"/check_availability?postal_code=" + postal_code, http.StatusFound)
					return
				} else {
					c.render(w, r, "failed_registration_attempt", data)
					return
				}
			} else {
				setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "error", "")), c.BootstrapAlertClass.Info)
				http.Redirect(w, r, r.URL.Host+"/check_availability?postal_code=" + postal_code, http.StatusFound)
				return
			}
		}
	}
}


func (c *AppController) AppIssueAction(w http.ResponseWriter, r *http.Request) {

	user_id := currentUserId(c.ControllerBase, r)

	if user_id == 0 {
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	u := models.FindUser(user_id)
	cj := models.ClientJurisdictionForUser(u, c.ControllerBase.PermissionGroups)
	jurisdiction_id := 0

	if cj == nil || len(cj.Jurisdictions) == 0 || cj.Jurisdictions[0].Id == 0 {
		http.Redirect(w, r, r.URL.Host+"/account", http.StatusFound)
		return
	}

	jurisdiction_id = cj.Jurisdictions[0].Id

	if r.Method == "GET" {
		data := models.AppIssue{
			JurisdictionId: jurisdiction_id,
			UserId: user_id,
		}
		c.render(w, r, "report_an_app_issue", data)
		return
	} else {
		data := models.AppIssue{
			JurisdictionId: jurisdiction_id,
			UserId: user_id,
			IssueType: r.FormValue("issue_type"),
			Description: r.FormValue("description"),
		}
		success := models.InsertAppIssue(&data)
		if success == true {
			//TODO SEND APP ISSUE EMAIL
			setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "request_has_been_submitted", "")), c.BootstrapAlertClass.Info)
			http.Redirect(w, r, r.URL.Host+"/account", http.StatusFound)
			return
		} else {
			c.render(w, r, "report_an_app_issue", data)
			return
		}
	}
}






func (c *AppController) RemoveAllStopsAction(w http.ResponseWriter, r *http.Request) {

	user_id := currentUserId(c.ControllerBase, r)
	if user_id == 0 {
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	query := `delete from bus_route_stop_users where user_id = $1`
	_, err := database.GetDB().Exec(query, user_id)
	if err != nil {
		setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r), "error_while_processing_request", "")), c.BootstrapAlertClass.Danger)
		http.Redirect(w, r, r.URL.Host+"/account", http.StatusFound)
		return
	}

	http.Redirect(w, r, r.URL.Host+"/", http.StatusFound)
}



func (c *AppController) ManageNotificationsAction(w http.ResponseWriter, r *http.Request) {

	user_id := currentUserId(c.ControllerBase, r)
	if(user_id == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	u := models.FindUser(user_id)

	cj := models.ClientJurisdictionForUser(u, c.ControllerBase.PermissionGroups)
	if cj == nil {
		http.Redirect(w, r, r.URL.Host+"/account", http.StatusFound)
		return
	}


	sns := models.ScanNotificationSubscriptionsForUser(user_id)

	data := struct {
		Jurisdictions *models.ClientJurisdictions
		NotificationSubscriptions *models.ScanNotificationSubscriptions
		HasNotificationSubscriptions bool
	} {
		cj,
		sns,
		(len(sns.Subscriptions) > 0),
	}

	c.render(w, r, "manage_notifications", data)
}





func (c *AppController) ManageSubscriptionsAction(w http.ResponseWriter, r *http.Request) {

	user_id := currentUserId(c.ControllerBase, r)
	if(user_id == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	u := models.FindUser(user_id)

	s := models.SubscriptionsForUser(u)
	if s == nil {
		http.Redirect(w, r, r.URL.Host+"/account", http.StatusFound)
		return
	}

	data := struct {
		Subscriptions *models.Subscriptions
	} {
		s,
	}

	c.render(w, r, "manage_subscriptions", data)
}


func (c *AppController) SubscriptionDetailsAction(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	user_id := currentUserId(c.ControllerBase, r)
	if(user_id == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	user := models.FindUser(user_id)

	subscription_id, err := strconv.Atoi(vars["subscription_id"])
	if err != nil {
		http.Redirect(w, r, r.URL.Host+"/account", http.StatusFound)
		return
	}

	subscription := models.FindSubscription(subscription_id)
	if subscription == nil {
		http.Redirect(w, r, r.URL.Host+"/account", http.StatusFound)
		return
	}

	jurisdiction := models.FindJurisdiction(subscription.JurisdictionId)
	if jurisdiction == nil {
		http.Redirect(w, r, r.URL.Host+"/account", http.StatusFound)
		return
	}

	sub_account_users := models.SubAccountUsersForSubscription(subscription_id)
	students := models.StudentsForUser(user.Id, jurisdiction.Id)
	scan_notification_subscriptions := models.ScanNotificationSubscriptionsForUser(user_id)



	data := struct {
		User *models.User
		SubAccountUsers *models.SubAccountUsers
		Subscription *models.Subscription
		Students *models.Students
		StudentCount int
		Jurisdiction *models.Jurisdiction
		ScanNotificationSubscriptions *models.ScanNotificationSubscriptions
		} {
		user,
		sub_account_users,
		subscription,
		students,
		len(students.StudentInformations),
		jurisdiction,
		scan_notification_subscriptions,
	}

	c.render(w, r, "subscription_details", data)
}










func (c *AppController) AddScanNotificationSubscriptionAction(w http.ResponseWriter, r *http.Request) {

	user_id := currentUserId(c.ControllerBase, r)
	if(user_id == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}


	jurisdiction_infos := r.Form["scan[][jurisdiction_id]"]
	scan_codes := r.Form["scan[][code]"]
	scan_names := r.Form["scan[][name]"]

	code_was_added := false

	for i := 0; i < len(jurisdiction_infos); i++ {
		jurisdiction_id, err := strconv.Atoi(jurisdiction_infos[i])
		if err != nil {
			continue
		}

		s := models.ScanNotificationSubscription{}
		s.Name = scan_names[i]
		s.Code = scan_codes[i]
		s.UserId = user_id
		s.JurisdictionId = jurisdiction_id
		sns_added := models.InsertScanNotificationSubscriptions(&s)
		if sns_added == true {
			code_was_added = true
		}
	}


	if code_was_added == false {
		setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "error_while_processing_request", "")), c.BootstrapAlertClass.Danger)
	}

	http.Redirect(w, r, r.URL.Host+ r.FormValue("out_action"), http.StatusFound)
}

func (c *AppController) RemoveScanNotificationSubscriptionAction(w http.ResponseWriter, r *http.Request) {

	user_id := currentUserId(c.ControllerBase, r)
	if(user_id == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	id := 0
	if len(r.FormValue("id")) > 0 {
		id, _ = strconv.Atoi(r.FormValue("id"))
	}

	deleted_scan_notification_subscription := models.DeleteScanNotificationSubscriptions(id)

	if deleted_scan_notification_subscription == false {
		setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "error_while_processing_request", "")), c.BootstrapAlertClass.Danger)
	}

	http.Redirect(w, r, r.URL.Host+ r.FormValue("out_action"), http.StatusFound)
}





func (c *AppController) AddStudentAction(w http.ResponseWriter, r *http.Request) {

	user_id := currentUserId(c.ControllerBase, r)
	if(user_id == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	subscription_id, err := strconv.Atoi(r.FormValue("subscription_id"))
	if err != nil {
		http.Redirect(w, r, r.URL.Host+"/account", http.StatusFound)
		return
	}

	student_identifier := r.FormValue("student_information[sis_identifier]")
	if len(student_identifier) == 0 {
		http.Redirect(w, r, r.URL.Host + "/subscription_details/" + r.FormValue("subscription_id") , http.StatusFound)
		return
	}

	student_added := models.AddStudentToSubscription(subscription_id, student_identifier)

	if student_added == false {
		setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "error_while_processing_request", "")), c.BootstrapAlertClass.Danger)
	}

	http.Redirect(w, r, r.URL.Host + "/subscription_details/" + r.FormValue("subscription_id") , http.StatusFound)
}














func (c *AppController) RemoveStudentAction(w http.ResponseWriter, r *http.Request) {

	user_id := currentUserId(c.ControllerBase, r)
	if(user_id == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	subscription_id, err := strconv.Atoi(r.FormValue("subscription_id"))
	if err != nil {
		http.Redirect(w, r, r.URL.Host+"/account", http.StatusFound)
		return
	}

	student_id, err := strconv.Atoi(r.FormValue("student_id"))
	if err != nil {
		setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "error_while_processing_request", "")), c.BootstrapAlertClass.Danger)
		http.Redirect(w, r, r.URL.Host + "/subscription_details/" + r.FormValue("subscription_id") , http.StatusFound)
	}

	deleted_student := models.RemoveStudentFromSubscription(subscription_id, student_id)

	if deleted_student == false {
		setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "error_while_processing_request", "")), c.BootstrapAlertClass.Danger)
	}

	http.Redirect(w, r, r.URL.Host + "/subscription_details/" + r.FormValue("subscription_id") , http.StatusFound)
}






func (c *AppController) AddSubAccountUserAction(w http.ResponseWriter, r *http.Request) {

	email := r.FormValue("email")
	password := r.FormValue("password")
	first_name := r.FormValue("first_name")
	last_name := r.FormValue("last_name")

	if len(email) == 0 {
		http.Redirect(w, r, r.URL.Host+"/account", http.StatusFound)
		return
	}

	user_id := currentUserId(c.ControllerBase, r)
	if(user_id == 0){
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	subscription_id, err := strconv.Atoi(r.FormValue("subscription_id"))
	if err != nil {
		http.Redirect(w, r, r.URL.Host+"/account", http.StatusFound)
		return
	}

	subscription := models.FindSubscription(subscription_id)
	if subscription == nil {
		http.Redirect(w, r, r.URL.Host+"/account", http.StatusFound)
		return
	}

	jurisdiction := models.FindJurisdiction(subscription.JurisdictionId)
	if jurisdiction == nil {
		http.Redirect(w, r, r.URL.Host + "/subscription_details/" + r.FormValue("subscription_id") , http.StatusFound)
		return
	}

	sub_account_users := models.SubAccountUsersForSubscription(subscription_id)
	if len(sub_account_users.Users) >= jurisdiction.SubAccountLimit {
		http.Redirect(w, r, r.URL.Host + "/subscription_details/" + r.FormValue("subscription_id") , http.StatusFound)
		return
	}

	sub_account_user_id := 0
	if models.EmailExists(email) {
		suid := models.UserIdForEmail(email)
		sub_account_user_id = suid
	} else {

		if len(password) == 0 || len(first_name) == 0 || len(last_name) == 0 {
			http.Redirect(w, r, r.URL.Host + "/subscription_details/" + r.FormValue("subscription_id") , http.StatusFound)
			return
		}

		suid, reg_err := models.RegisterUser(email, password, first_name, last_name, c.PermissionGroups.SubAccount)
		if reg_err != nil {
			setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  reg_err.Error(), "")), c.BootstrapAlertClass.Danger)
			http.Redirect(w, r, r.URL.Host + "/subscription_details/" + r.FormValue("subscription_id") , http.StatusFound)
		}
		sub_account_user_id = suid
	}

	if sub_account_user_id == 0 {
		setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "error_while_processing_request", "")), c.BootstrapAlertClass.Danger)
		http.Redirect(w, r, r.URL.Host + "/subscription_details/" + r.FormValue("subscription_id") , http.StatusFound)
	}

	permission_group_added := models.AddPermissionGroupToUser(sub_account_user_id, c.PermissionGroups.SubAccount)
	if permission_group_added == false {
		setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "error_while_processing_request", "")), c.BootstrapAlertClass.Danger)
		http.Redirect(w, r, r.URL.Host + "/subscription_details/" + r.FormValue("subscription_id") , http.StatusFound)
	}


	sub_account_added := models.AddSubAccountUserToSubscription(subscription, sub_account_user_id)
	if sub_account_added == false {
		setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "error_while_processing_request", "")), c.BootstrapAlertClass.Danger)
	}

	http.Redirect(w, r, r.URL.Host + "/subscription_details/" + r.FormValue("subscription_id") , http.StatusFound)
}












func (c *AppController) LostItemReportAction(w http.ResponseWriter, r *http.Request) {

	user_id := currentUserId(c.ControllerBase, r)

	if user_id == 0 {
		http.Redirect(w, r, r.URL.Host+"/login", http.StatusFound)
		return
	}

	u := models.FindUser(user_id)
	cj := models.ClientJurisdictionForUser(u, c.PermissionGroups)
	if cj == nil {
		http.Redirect(w, r, r.URL.Host+"/account", http.StatusFound)
		return
	}

	jurisdiction_id := 0
	for i := 0; i < len(cj.Jurisdictions); i++ {
		if cj.Jurisdictions[i].HasLostItemReports == true {
			jurisdiction_id = cj.Jurisdictions[i].Id
			break
		}
	}
	if jurisdiction_id == 0 {
		http.Redirect(w, r, r.URL.Host+"/account", http.StatusFound)
		return
	}



	if r.Method == "GET" {
		data := models.LostItemReport{}
		data.Email = u.Email
		data.JurisdictionId = jurisdiction_id
		c.render(w, r, "lost_item_report", data)
		return

	} else {

		data := models.LostItemReport{
			JurisdictionId: jurisdiction_id,
			FirstName: r.FormValue("first_name"),
			LastName: r.FormValue("last_name"),
			Email: u.Email,
			Summary: r.FormValue("summary"),
			Phone: r.FormValue("phone"),
			RouteIdentifier: r.FormValue("route_identifier"),
			DateLost: r.FormValue("date_lost"),
			Description: r.FormValue("description"),
		}
		success := models.InsertLostItemReport(&data)
		if success == true {
			//TODO SEND LOST ITEM REPORT EMAIL
			setFlash(c.ControllerBase, r, w, string(T(currentLocale(c.ControllerBase, r),  "request_has_been_submitted", "")), c.BootstrapAlertClass.Info)
			http.Redirect(w, r, r.URL.Host+"/account", http.StatusFound)
			return
		} else {
			c.render(w, r, "lost_item_report", data)
			return
		}

	}

}



















// Redirects
//----------------------------------------------------------------------------------------------------------------------

func (c *AppController) redirectToJoinIfNotALoggedIn(w http.ResponseWriter, r *http.Request) {
	session, _ :=  c.SessionStore.Get(r, "auth")
	email := session.Values["current_user_email"]
	if email == nil {
		http.Redirect(w, r, r.URL.Host+"/join", http.StatusFound)
		return
	}
}

func (c *AppController) redirectToLoginIfNotLoggedIn(w http.ResponseWriter, r *http.Request) {
	session, _ :=  c.SessionStore.Get(r, "auth")
	email := session.Values["current_user_email"]
	if email == nil {
		http.Redirect(w, r, r.URL.Host+"/join", http.StatusFound)
		return
	}
}

// Helpers

func (c *AppController) getCurrentUserEmail(r *http.Request) string {
	session, _ :=  c.SessionStore.Get(r, "auth")
	email := session.Values["current_user_email"]
	return email.(string)
}

func (c *AppController) getCurrentUser(r *http.Request) *models.User {
	email := c.getCurrentUserEmail(r)
	user := models.FindUserByEmail(email)
	return user
}



//func validateToken(token string) models.AuthInfo {
//	a := models.AuthInfo{}
//	u := models.FindUserByToken(token)
//	if(u != nil){
//		a.User = u
//		a.TokenValid = true
//	}
//	return a
//}


// addAction requires you to have a view named <action>.html and a method func (c *AppController) <Action>Action(http.ResponseWriter, *http.Request)
func (c *AppController) addAction(action string){
	//TODO: determine if this can be moved to ControllerBase and if c can just be cast to the correct type.
	//fmt.Println(strings.Title(action)+"Action")
	c.addTemplateApp(action)
	c.Router.HandleFunc(c.RoutePrefix+"/"+action, reflect.ValueOf(c).MethodByName(strings.Title(action)+"Action").Interface().(func(http.ResponseWriter, *http.Request)))
}