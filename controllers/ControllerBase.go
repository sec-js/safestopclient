package controllers

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/schoolwheels/safestopclient/i18n"
	"github.com/schoolwheels/safestopclient/models"
	"github.com/spf13/viper"
	"html/template"
	"log"
	"net/http"
	"os"
	"runtime"
)




type ControllerBase struct {
	Name string
	RoutePrefix string
	Templates map[string]*template.Template
	Router *mux.Router
	SessionStore *sessions.CookieStore
	BootstrapAlertClass *models.BootstrapAlertClass
	PermissionGroups *models.PermissionGroups
}



func funcName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

//TODO: create addAction method that maps a route to a function directly with templates and with intelligent defaults.

func (c *ControllerBase) addRouteWithPrefix(route string, handler func(http.ResponseWriter, *http.Request) ){
	c.Router.HandleFunc(c.RoutePrefix+route, handler)
}

func (c *ControllerBase) addRouteWithPrefixMethod(route string, handler func(http.ResponseWriter, *http.Request), method string ){
	c.Router.HandleFunc(c.RoutePrefix+route, handler).Methods(method)
}

func (c *ControllerBase) addTemplateNoNav(name string){
	c.addTemplate(name, name + ".html", "no_nav.html")
}

func (c *ControllerBase) addTemplateApp(name string){
	c.addTemplate(name, name + ".html", "app.html")
}



func (c *ControllerBase) SendEmail(r *http.Request, to []string, subject string,  template_name string, viewModel interface{}) bool {


	file_path := "views/mail/" + template_name + ".html"
	if _, err := os.Stat(file_path); os.IsNotExist(err) {
		log.Println(err)
		return false
	}

	var currentUser CurrentUser
	session, _ :=  c.SessionStore.Get(r, "auth")
	if session.Values["current_user_email"] != nil {
		currentUser = CurrentUser {Email: session.Values["current_user_email"].(string)}
	}

	var data = ViewModel {
		CurrentUser:  currentUser,
		CurrentUserId: currentUserId(c,r),
		CurrentLocale: currentLocale(c,r),
		Domain: viper.GetString("domain"),
		SupportNumber: viper.GetString("support_number"),
		ViewData: viewModel,
	}

	m := models.NewMailRequest(to, subject)

	funcMap := template.FuncMap{"t": T}
	t := template.Must(template.New(template_name + ".html").Funcs(funcMap).ParseFiles("views/mail/" + template_name + ".html"))

	if t == nil {
		return false
	}

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		log.Println(err)
		return false
	}
	m.Body = buf.String()

	ok, _ := m.SendEmail()
	return ok
}




func (c *ControllerBase) addTemplate(name string, file string, layout string){
	if c.Templates == nil {
		c.Templates = make(map[string]*template.Template)
	}

	funcMap := template.FuncMap{"mod": mod, "n": N, "t": T}

	c.Templates[name] = template.Must(template.New("base.html").Funcs(funcMap).ParseFiles("views/"+c.Name+"/"+file, "views/layouts/"+layout, "views/layouts/base.html"))
}

func (c *ControllerBase) renderJSON(data interface{}, w http.ResponseWriter) {

	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (c *ControllerBase) render(w http.ResponseWriter, r *http.Request, template string, data interface{} ) {
	c.renderTemplate(w, r, template, "layout", data)
}

func T(locale string, key string, value string, args ...interface{}) template.HTML {
	return i18n.GetI18n().Default(value).T(locale, key, args...)
}




func mod(i, j int) bool {
	return i%j == 0
}

func N(start, end int) (stream chan int) {
	stream = make(chan int)
	go func() {
		for i := start; i <= end; i++ {
			stream <- i
		}
		close(stream)
	}()
	return
}

type CurrentUser struct {
	Email string
}




type FlashMessages struct {
	FlashMessages []FlashMessage
}

type FlashMessage struct {
	Message string
	BootstrapClass string
}

//append session data
type ViewModel struct {
	FlashMessages FlashMessages
	CurrentUser CurrentUser
	CurrentUserId int
	CurrentLocale string
	CurrentPath string
	Domain string
	SupportNumber string
	CSRFTemplateField template.HTML
	ViewData interface{}
}



func (c *ControllerBase) renderTemplate(w http.ResponseWriter, r *http.Request, name string, template_name string, viewModel interface{}) {

	var currentUser CurrentUser
	session, _ :=  c.SessionStore.Get(r, "auth")
	if session.Values["current_user_email"] != nil {
		currentUser = CurrentUser {Email: session.Values["current_user_email"].(string)}
	}

	var data = ViewModel{
		CurrentUser:  currentUser,
		CurrentUserId: currentUserId(c,r),
		CurrentLocale: currentLocale(c,r),
		Domain: viper.GetString("domain"),
		SupportNumber: viper.GetString("support_number"),
		CurrentPath: r.URL.Path,
		ViewData: viewModel,
		CSRFTemplateField: csrf.TemplateField(r),
	}
	getFlash(c,r, w, &data)


	// Ensure the template exists in the map.
	tmpl, ok := c.Templates[name]
	if !ok {
		http.Error(w, "The template does not exist.", http.StatusInternalServerError)
	}


	err := tmpl.ExecuteTemplate(w, template_name, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}




func currentLocale(c *ControllerBase, r *http.Request) string {
	session, _ :=  c.SessionStore.Get(r, "auth")
	locale := session.Values["locale"]
	if(locale == nil || locale == ""){
		locale = "en"
	}
	return locale.(string)
}


func setCurrentUserId(c *ControllerBase, r *http.Request, w http.ResponseWriter, id int) {
	session, _ :=  c.SessionStore.Get(r, "auth")
	session.Values["user_id"] = id
	session.Save(r, w)
}

func currentUserId(c *ControllerBase, r *http.Request) int {
	session, _ :=  c.SessionStore.Get(r, "auth")
	user_id := session.Values["user_id"]
	if(user_id == nil || user_id == ""){
		user_id = 0
	}
	return user_id.(int)
}




func setFlash(c *ControllerBase, r *http.Request, w http.ResponseWriter, message string , bootstrap_class string ){
	session, _ :=  c.SessionStore.Get(r, "flash")
	f := FlashMessage{ Message: message, BootstrapClass: bootstrap_class }
	session.AddFlash(f, "message")
	session.Save(r, w)
}

func getFlash(c *ControllerBase, r *http.Request, w http.ResponseWriter, data *ViewModel){
	session, _ :=  c.SessionStore.Get(r, "flash")
	data.FlashMessages.FlashMessages = []FlashMessage{}
	messages := session.Flashes("message");
	if len(messages) > 0 {
		for i := 0; i < len(messages); i++ {
			data.FlashMessages.FlashMessages = append(data.FlashMessages.FlashMessages, messages[i].(FlashMessage))
		}
	}
	session.Save(r, w)
}




