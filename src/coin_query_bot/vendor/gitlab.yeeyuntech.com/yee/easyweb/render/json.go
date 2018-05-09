/**
 * Created by angelina on 2017/8/30.
 */

package render

import (
	"net/http"
	"encoding/json"
)

var jsonContentType = []string{"application/json; charset=utf-8"}

type JSON struct {
	Data interface{}
}

func (j JSON) Render(w http.ResponseWriter) error {
	j.WriteContentType(w)
	jsonBytes, err := json.Marshal(j.Data)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	if err != nil {
		return err
	}
	return nil
}

func (j JSON) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}

type IndentedJSON struct {
	Data interface{}
}

func (ij IndentedJSON) Render(w http.ResponseWriter) error {
	ij.WriteContentType(w)
	jsonBytes, err := json.MarshalIndent(ij.Data, "", "		")
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	if err != nil {
		return err
	}
	return nil
}

func (ij IndentedJSON) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}
