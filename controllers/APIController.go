package controllers

import (
	"net/http"
	"encoding/json"

)

type APIController struct {
	*ControllerBase
}

func (c *APIController) Register() {

	c.RoutePrefix = "/api"

	//actions
	c.addRouteWithPrefix("/version", c.versionAction)
}

func (c *APIController) versionAction(w http.ResponseWriter, r *http.Request) {

	data := struct {
		Version float64 `json:"version"`
	}{
		1.0,
	}

	c.renderJSON(data, w)
}

func (c *APIController) renderJSON(data interface{}, w http.ResponseWriter) {

	js, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}


