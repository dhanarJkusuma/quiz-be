package util

import (
	"encoding/json"
	"github.com/dhanarJkusuma/pager"
	"net/http"
)

func HandleJson(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

type Util struct {
	authModule *pager.Auth
}
