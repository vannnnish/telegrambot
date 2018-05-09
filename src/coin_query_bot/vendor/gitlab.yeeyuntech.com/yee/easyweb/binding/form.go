/**
 * Created by angelina on 2017/8/28.
 */

package binding

import "net/http"

type formBinding struct{}

var _ Binding = &formBinding{}

func (formBinding) Name() string {
	return "form"
}

func (formBinding) Bind(req *http.Request, obj interface{}) error {
	if err := req.ParseForm(); err != nil {
		return err
	}
	req.ParseMultipartForm(32 << 10)
	if err := mapForm(obj, req.Form); err != nil {
		return err
	}
	err := validate(obj)
	if err != nil {
		return err
	}
	return nil
}
