package render

import (
    "encoding/json"
    "net/http"
)

type (
    JSONP struct {
        Data     interface{}
        Callback string
    }
)

var jsonpContentType = []string{"text/plain; charset=utf-8"}

func (r JSONP) Render(w http.ResponseWriter) error {
    return WriteJSONP(w, r.Data, r.Callback)
}

func WriteJSONP(w http.ResponseWriter, obj interface{}, callback string) error {
    writeContentType(w, jsonContentType)
    w.Write([]byte(callback))
    w.Write([]byte("("))
    err := json.NewEncoder(w).Encode(obj)
    w.Write([]byte(")"))
    return err
}

