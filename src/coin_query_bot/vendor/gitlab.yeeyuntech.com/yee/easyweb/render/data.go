/**
 * Created by angelina on 2017/8/30.
 */

package render

import "net/http"

type Data struct {
	ContentType string
	Data        []byte
}

func (d Data) Render(w http.ResponseWriter) error {
	d.WriteContentType(w)
	_, err := w.Write(d.Data)
	return err
}

func (d Data) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, []string{d.ContentType})
}
