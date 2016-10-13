package assert

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func readBody(res *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}
	// Re-fill body reader stream after reading it
	res.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return body, err
}

// BodyMatchString asserts a response body matching a string expression.
// Regular expressions can be used as value to perform the specific assertions.
func BodyMatchString(pattern string) Func {
	return func(res *http.Response, req *http.Request) error {
		body, err := readBody(res)
		if err != nil {
			return err
		}
		if match, _ := regexp.MatchString(pattern, string(body)); !match {
			return fmt.Errorf("Body mismatch: pattern not found '%s'", pattern)
		}
		return nil
	}
}

// BodyEquals asserts as strict equality comparison the
// response body with a given string string.
func BodyEquals(value string) Func {
	return func(res *http.Response, req *http.Request) error {
		body, err := readBody(res)
		if err != nil {
			return err
		}

		bodyStr := string(body)
		err = fmt.Errorf("Bodies mismatch:\n\thave: %#v\n\twant: %#v\n", bodyStr, value)

		// Remove line feed sequence
		if len(bodyStr) > 0 && bodyStr[len(bodyStr)-1] == '\n' {
			bodyStr = bodyStr[:len(bodyStr)-1]
		}

		// Perform the comparison
		if len(bodyStr) != len(value) || value != bodyStr {
			return err
		}

		return nil
	}
}

// BodyLength asserts a response body length.
func BodyLength(length int) Func {
	return func(res *http.Response, req *http.Request) error {
		cl, err := strconv.Atoi(res.Header.Get("Content-Length"))
		// Infer length from body buffer
		if err != nil || cl == 0 {
			body, err := readBody(res)
			if err != nil {
				return err
			}
			cl = len(body)
		}
		// Match body length
		if cl != length {
			return fmt.Errorf("Body length mismatch: '%d' should be equal to '%d'", cl, length)
		}
		return nil
	}
}

func BodySnap(name string, ignoredFields map[string]FieldFunc) Func {
	return func(res *http.Response, req *http.Request) error {
		url := req.URL.String()

		body, err := readBody(res)
		if err != nil {
			return err
		}

		deleteFields := func(item map[string]interface{}) error {
			for path, _ := range ignoredFields {
				var value interface{} = item
				var v map[string]interface{}
				var ok bool
				var part string
				parts := strings.Split(path, ".")
				for _, part = range parts {
					if v, ok = value.(map[string]interface{}); ok {
						value = v[part]
					} else {
						return errors.New("couldn't find item at path: " + path)
					}
				}
				delete(v, part)
			}
			return nil
		}

		var data interface{}
		json.Unmarshal(body, &data)

		switch d := data.(type) {
		case []interface{}:
			for _, item := range d {
				if v, ok := item.(map[string]interface{}); ok {
					err := deleteFields(v)
					if err != nil {
						return err
					}
				}
			}
		case map[string]interface{}:
			err := deleteFields(d)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("Unrecognized type for Field in body snap: %T", d)
		}

		success, diffString, err := Shot(name+"-body", data)
		if err != nil {
			return err
		}
		if !success {
			return NewSnapError(fmt.Sprintf("Body test for url '%s' failed.\n See this diff for more detail:\n\n %s", url, diffString))
		}
		return nil
	}
}
