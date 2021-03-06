package assert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
)

func unmarshal(buf []byte) (map[string]interface{}, error) {
	data := make(map[string]interface{})
	if err := json.Unmarshal(buf, &data); err != nil {
		return data, fmt.Errorf("failed to Unmarshal: %s", err)
	}
	return data, nil
}

func unmarshalBody(res *http.Response) (map[string]interface{}, error) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, nil
	}
	return unmarshal(body)
}

func marshal(data map[string]interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func readBodyJSON(res *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}

	// Re-fill body reader stream after reading it
	res.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return body, err
}

func compare(body map[string]interface{}, data interface{}) error {
	var err error
	var match map[string]interface{}

	// Infer and cast string input
	if jsonStr, ok := data.(string); ok {
		data = []byte(jsonStr)
	}

	// Infer and cast string input
	if reader, ok := data.(io.Reader); ok {
		data, err = ioutil.ReadAll(reader)
		if err != nil {
			return err
		}
	}

	// Infer and cast input as map
	if fields, ok := data.(map[string]interface{}); ok {
		data, err = marshal(fields)
		if err != nil {
			return err
		}
	}

	// Infer and cast bytes input
	buf, ok := data.([]byte)
	if ok {
		match, err = unmarshal(buf)
		if err != nil {
			return err
		}
	}

	// Assert via string
	bodyBuf, err := marshal(body)
	if err != nil {
		return err
	}
	matchBuf, err := marshal(match)
	if err != nil {
		return err
	}

	// Compare by byte sequences
	if !reflect.DeepEqual(bodyBuf, matchBuf) {
		return fmt.Errorf("JSON mismatch:\n\thave: %#v\n\twant: %#v\n", string(bodyBuf), string(matchBuf))
	}

	return nil
}

// JSON deeply and strictly compares the JSON
// response body with the given JSON structure.
func JSON(data interface{}) Func {
	return func(res *http.Response, req *http.Request) error {
		// Read and unmarshal response body as JSON
		body, err := unmarshalBody(res)
		if err != nil {
			return err
		}

		return compare(body, data)
	}
}
