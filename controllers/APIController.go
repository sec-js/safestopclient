package controllers

import (
	"net/http"
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


