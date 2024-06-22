package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

// page this is use for template
type page struct {
	Title             string
	CSRFToken         string
	ErrorMessages     []string
	ErrorMessagesJSON string
	Data              interface{}
	DataJSON          string
	Fullname          string
	Username          string
	Roles             []string
	UIMapData         map[string]interface{}
}

// NewPage ...
func NewPage() *page {
	return new(page)
}

// AddError ...
func (p *page) AddError(msg string) {
	p.ErrorMessages = append(p.ErrorMessages, msg)
	p.ErrorMessagesJSON = p.justJSONMarshal(p.ErrorMessages)
}

// ResetErrors ...
func (p *page) ResetErrors(msg string) {
	p.ErrorMessages = nil
	p.ErrorMessages = make([]string, 0)
	p.ErrorMessagesJSON = p.justJSONMarshal(p.ErrorMessages)
}

// SetData ...
func (p *page) SetData(v interface{}) {
	p.Data = v
	p.DataJSON = p.justJSONMarshal(p.Data)
}

// JSONify ...
func (p *page) JSONify() string {
	p.DataJSON = p.justJSONMarshal(p.Data)
	return p.justJSONMarshal(p)
}

func (p *page) justJSONMarshal(v interface{}) string {
	result, err := json.Marshal(v)
	if err != nil {
		log.Panic(err)
	}
	return string(result)
}

func (p *page) RenderPage(w http.ResponseWriter, r *http.Request, filenames ...string) {
	if p.Data == nil {
		p.SetData(make(map[string]interface{}))
	}

	if p.ErrorMessages == nil {
		p.ResetErrors("")
	}

	t, err := template.New("base").ParseFS(viewsFS, filenames...)
	if err != nil {
		log.Panic(err)
		return
	}

	err = t.Execute(w, p)
	if err != nil {
		log.Panic(err)
	}
}
