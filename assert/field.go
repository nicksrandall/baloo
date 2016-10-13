package assert

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func BuildFieldTest(ignoredFields map[string]FieldFunc) Func {
	return func(res *http.Response, req *http.Request) error {
		body, err := readBody(res)
		if err != nil {
			return err
		}

		testFields := func(item map[string]interface{}) error {
			for path, fn := range ignoredFields {
				var value interface{} = item
				parts := strings.Split(path, ".")
				for _, part := range parts {
					if v, ok := value.(map[string]interface{}); ok {
						value = v[part]
					} else {
						return errors.New("couldn't find item at path: " + path)
					}
				}
				err := fn(value)
				if err != nil {
					return err
				}
			}
			return nil
		}

		var data interface{}
		json.Unmarshal(body, &data)

		// if array and contains map then use it for fields
		switch d := data.(type) {
		case []interface{}:
			for _, item := range d {
				if v, ok := item.(map[string]interface{}); ok {
					err := testFields(v)
					if err != nil {
						return err
					}
				}
			}
			return nil
		case map[string]interface{}:
			err := testFields(d)
			if err != nil {
				return err
			}
			return nil
		default:
			return fmt.Errorf("Unrecognized type for Field in field test: %T", d)
		}
	}
}

func IgnoreField(v interface{}) error {
	return nil
}

func FieldIsNumber(v interface{}) error {
	if _, ok := v.(float64); ok {
		return nil
	}
	return errors.New(fmt.Sprintf("%#v is not a number", v))
}

func FieldIsBool(v interface{}) error {
	if _, ok := v.(bool); ok {
		return nil
	}
	return errors.New(fmt.Sprintf("%#v is not a boolean", v))
}

func FieldIsString(v interface{}) error {
	if _, ok := v.(string); ok {
		return nil
	}
	return errors.New(fmt.Sprintf("%#v is not a string", v))
}
