package main

import (
	"encoding/json"
	"fmt"
	"github.com/rs/xid"
	"math/rand"
	"net/http"
	"time"
)

func generateRandomBytes(length int) []byte {
	s := ""
	for i := 33; i <= 126; i++ {
		s = s + fmt.Sprintf("%c", i)
	}
	rs := make([]byte, 0)
	randSource := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		delta := randSource.Intn(len(s))
		rs = append(rs, s[delta])
	}
	return rs
}

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
