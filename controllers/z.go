package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/gorilla/sessions"
	"github.com/johnpili/http-probe/models"
	"github.com/psi-incontrol/go-sprocket/page"
	"github.com/psi-incontrol/go-sprocket/sprocket"
	"github.com/rs/xid"
)

var (
	cookieStore   *sessions.CookieStore
	viewBox       *rice.Box
	staticBox     *rice.Box
	configuration *models.Config
)

//New ...
func New(vb *rice.Box, sb *rice.Box, store *sessions.CookieStore, config *models.Config) *Hub {
	viewBox = vb
	staticBox = sb
	cookieStore = store
	configuration = config
	return new(Hub)
}

//Hub ...
type Hub struct {
	PageController PageController
}

func renderPage(w http.ResponseWriter, r *http.Request, vm interface{}, filenames ...string) {
	page := vm.(*page.Page)

	if page.Data == nil {
		page.SetData(make(map[string]interface{}))
	}

	if page.ErrorMessages == nil {
		page.ResetErrors("")
	}

	cookieName := os.Getenv(configuration.System.EnvCookieName)
	session, err := cookieStore.Get(r, cookieName)
	if err == nil {
		if session.Values["roles"] == nil {
			page.Roles = []string{}
		} else {
			page.Roles = session.Values["roles"].([]string)
		}

		if session.Values["username"] == nil {
			page.Username = ""
		} else {
			page.Username = session.Values["username"].(string)
		}

		if session.Values["fullname"] == nil {
			page.Fullname = ""
		} else {
			page.Fullname = session.Values["fullname"].(string)
		}
	}

	x, err := sprocket.GetTemplates(viewBox, filenames)
	err = x.Execute(w, page)
	if err != nil {
		log.Panic(err.Error())
	}
}

func respondWithJSON(w http.ResponseWriter, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(200)
	w.Write(response)
}

func justJSONMarshal(v interface{}) string {
	result, err := json.Marshal(v)
	if err != nil {
		log.Panic(err)
	}
	return string(result)
}

func metaValueExtractor(v interface{}) string {
	if v != nil {
		return v.(string)
	}
	return ""
}

func interfaceArrayToStringArray(t []interface{}) []string {
	s := make([]string, len(t))
	for i, v := range t {
		s[i] = fmt.Sprint(v)
	}
	return s
}

func parseToDatabaseDate(v string) interface{} {
	dateLayout := `02/01/2006`
	parsedDate, err := time.Parse(dateLayout, strings.TrimSpace(v))
	if err != nil {
		return nil
	}
	return parsedDate
}

func generateGUID() string {
	guid := xid.New()
	guid.Time().Add(100)
	return guid.String()[12:20]
}
