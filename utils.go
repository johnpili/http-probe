package main

import (
	"encoding/json"
	"fmt"
	"github.com/rs/xid"
	"net/http"
)

func generateGUID() string {
	guid := xid.New()
	guid.Time().Add(100)
	return guid.String()[12:20]
}

// respondStatusCodeWithJSON ...
func respondStatusCodeWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	response, err := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		crashPayload := []byte(`{
			"success": false,
			"result": null,
			"errors": ["` + fmt.Sprintf("%s", err.Error()) + `"]
		}`)
		w.Write(crashPayload)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(response)
}
