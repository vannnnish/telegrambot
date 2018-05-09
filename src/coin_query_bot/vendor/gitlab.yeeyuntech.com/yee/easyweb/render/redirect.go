/**
 * Created by angelina on 2017/8/30.
 */

package render

import (
	"net/http"
	"fmt"
)

type Redirect struct {
	Request *http.Request
	Code    int
	Url     string
}

func (r Redirect) Render(w http.ResponseWriter) error {
	if (r.Code < 300 || r.Code > 308) && r.Code != 201 {
		panic(fmt.Sprintf("Cannot redirect with status code %d", r.Code))
	}
	http.Redirect(w, r.Request, r.Url, r.Code)
	return nil
}

func (r Redirect) WriteContentType(w http.ResponseWriter) {
}
