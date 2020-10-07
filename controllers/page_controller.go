package controllers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-zoo/bone"

	"github.com/psi-incontrol/go-sprocket/page"
)

//PageController ...
type PageController struct{}

// DashboardHandler ...
func (z *PageController) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	cookieName := os.Getenv(configuration.System.EnvCookieName)
	session, _ := cookieStore.Get(r, cookieName)

	id := bone.GetValue(r, "id")
	if len(id) == 0 {
		if id, ok := session.Values["id"].(string); ok {
			http.Redirect(w, r, fmt.Sprintf("/%s", id), 302)
			return
		}

		id = generateGUID()
		session.Values["id"] = id
		session.Save(r, w)

		http.Redirect(w, r, fmt.Sprintf("/%s", id), 302)
		return
	}

	deltaID1 := ""
	if d, ok := session.Values["id"].(string); ok {
		deltaID1 = d
	}

	if deltaID1 != id { // Need to check if someone is using an ID that don't belong to their session
		id = generateGUID()
		session.Values["id"] = id
		session.Save(r, w)

		http.Redirect(w, r, fmt.Sprintf("/%s", id), 302)
		return
	}

	data := make(map[string]interface{})
	data["room"] = id

	page := page.New()
	page.Title = "http-probe | probe.johnpili.com"
	page.SetData(data)

	renderPage(w, r, page, "base.html", "index.html")
}

// DocumentationHandler ...
func (z *PageController) DocumentationHandler(w http.ResponseWriter, r *http.Request) {
	page := page.New()
	page.Title = "Documentation | "

	renderPage(w, r, page, "base.html", "documentation.html")
}
