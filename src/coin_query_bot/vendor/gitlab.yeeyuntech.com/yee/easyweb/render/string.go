/**
 * Created by angelina on 2017/8/30.
 */

package render

import (
	"net/http"
	"fmt"
	"io"
)

var plainContentType = []string{"text/plain; charset=utf-8"}

type String struct {
	Format string
	Data   []interface{}
}

func (s String) Render(w http.ResponseWriter) error {
	s.WriteContentType(w)
	var err error
	if len(s.Data) > 0 {
		_, err = fmt.Fprintf(w, s.Format, s.Data...)
	} else {
		_, err = io.WriteString(w, s.Format)
	}
	return err
}

func (s String) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, plainContentType)
}
