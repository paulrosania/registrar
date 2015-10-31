package main

import (
	"fmt"

	"encoding/json"
	"net/http"
)

type OAuthError struct {
	Type        string `json:"error"`
	Description string `json:"error_description"`

	Meta map[string]interface{} `json:"meta,omitempty"`
}

func NewOAuthError(typ, desc string) *OAuthError {
	return &OAuthError{
		Type:        typ,
		Description: desc,

		Meta: make(map[string]interface{}),
	}
}

func (e *OAuthError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Description)
}

func (e *OAuthError) Write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch e.Type {
	case "access_denied":
		w.WriteHeader(http.StatusUnauthorized)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}

	out, err := json.Marshal(e)
	if err != nil {
		fmt.Fprintf(w, `{error:"server_error"}\n`)
		return
	}

	w.Write(out)
	w.Write([]byte("\n"))
}
