/**
 * Created by angelina on 2017/8/28.
 */

package binding

import (
	"net/http"
	"sync"
	"errors"
	"fmt"
	"gitlab.yeeyuntech.com/yee/easyweb/validation"
	"strings"
)

const (
	MIMEJSON              = "application/json"
	MIMEHTML              = "text/html"
	MIMEXML               = "application/xml"
	MIMEXML2              = "text/xml"
	MIMEPlain             = "text/plain"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
)

var (
	Form = formBinding{}
)

type Binding interface {
	Name() string
	Bind(*http.Request, interface{}) error
}

func Default(method, contentType string) Binding {
	if method == "GET" {
		return Form
	}
	switch contentType {
	case MIMEJSON:
		// todo
	case MIMEXML, MIMEXML2:
		// todo
	default: //case MIMEPOSTForm, MIMEMultipartPOSTForm:
		return Form
	}
	return Form
}

var (
	defaultValidation validation.Validation
	once              sync.Once
)

func initValidation() {
	once.Do(func() {
		defaultValidation = validation.Validation{}
	})
}

func validate(obj interface{}) error {
	initValidation()
	defaultValidation.Clear()
	ok, err := defaultValidation.Valid(obj)
	if err != nil {
		return err
	}
	if !ok && defaultValidation.HasErrors() {
		msg := fmt.Sprintf("参数%s错误:%s", firstLetterToLower(defaultValidation.Errors[0].Field),
			defaultValidation.Errors[0].Message)
		return errors.New(msg)
	}
	return nil
}

func firstLetterToLower(s string) string {
	if len(s) == 0 {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}
