package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-zoo/bone"

	"github.com/psi-incontrol/go-sprocket/page"
)

// PageController ...
type PageController struct{}

// RequestMapping ...
func (z *PageController) RequestMapping(router *bone.Mux) {
	router.GetFunc("/documentation", z.DocumentationHandler)
	router.GetFunc("/", z.DashboardHandler)
	router.GetFunc("/:id", z.DashboardHandler)
}

// DashboardHandler ...
func (z *PageController) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	session, err := cookieStore.Get(r, configuration.System.CookieName)
	if err != nil {
		log.Fatal(err)
	}

	id := bone.GetValue(r, "id")
	if len(id) == 0 {
		if id, ok := session.Values["id"].(string); ok {
			http.Redirect(w, r, fmt.Sprintf("/%s", id), 302)
			return
		}

		log.Println("Generating")
		generatedID := generateGUID()
		session.Values["id"] = generatedID
		session.Save(r, w)

		http.Redirect(w, r, fmt.Sprintf("/%s", generatedID), 302)
		return
	}

	extractedID := ""
	if e, ok := session.Values["id"].(string); ok {
		extractedID = e
	}

	if len(extractedID) == 0 {
		http.Redirect(w, r, fmt.Sprintf("/"), 302)
		return
	}

	if strings.Compare(id, extractedID) != 0 { // Need to check if someone is using an ID that don't belong to their session
		http.Redirect(w, r, fmt.Sprintf("/%s", extractedID), 302)
		return
	}

	data := make(map[string]interface{})
	data["room"] = id

	page := page.New()
	page.Title = "An HTTP probe tool with web interface | probe.johnpili.com"
	page.SetData(data)

	renderPage(w, r, page, "base.html", "index.html")
}

// DocumentationHandler ...
func (z *PageController) DocumentationHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := cookieStore.Get(r, configuration.System.CookieName)

	data := make(map[string]interface{})
	if id, ok := session.Values["id"].(string); ok {
		data["room"] = id
	}

	page := page.New()
	page.Title = "Documentation | probe.johnpili.com"
	page.SetData(data)

	renderPage(w, r, page, "base.html", "documentation.html")
}
