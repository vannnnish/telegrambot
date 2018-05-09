/**
 * Created by angelina on 2017/8/29.
 */

package render

import "net/http"

type Render interface {
	Render(http.ResponseWriter) error
	WriteContentType(http.ResponseWriter)
}

var (
	_ Render = &JSON{}
	_ Render = &IndentedJSON{}
	_ Render = &String{}
	_ Render = &Data{}
	_ Render = &Redirect{}
	_ Render = &HTML{}
)

func writeContentType(w http.ResponseWriter, values []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = values
	}
}
