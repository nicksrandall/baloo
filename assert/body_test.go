package assert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/nbio/st"
)

func TestBodyMatchString(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewBufferString("hello world"))
	res := &http.Response{Body: body}
	st.Expect(t, BodyMatchString("hello")(res, nil), nil)
	st.Expect(t, BodyMatchString("^hello world$")(res, nil), nil)
	st.Expect(t, BodyMatchString("world$")(res, nil), nil)
	st.Expect(t, BodyMatchString("he[a-z]+")(res, nil), nil)
	st.Reject(t, BodyMatchString("foo")(res, nil), nil)
	st.Reject(t, BodyMatchString("bar")(res, nil), nil)
}

func TestBodyLength(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewBufferString("hello world"))
	res := &http.Response{Body: body}
	st.Expect(t, BodyLength(11)(res, nil), nil)
	st.Reject(t, BodyLength(10)(res, nil), nil)
	st.Reject(t, BodyLength(0)(res, nil), nil)
}

func TestBodyEquals(t *testing.T) {
	body := ioutil.NopCloser(bytes.NewBufferString("hello world"))
	res := &http.Response{Body: body}
	st.Expect(t, BodyEquals("hello world")(res, nil), nil)
	st.Reject(t, BodyEquals("hello")(res, nil), nil)
	st.Reject(t, BodyEquals("world")(res, nil), nil)
	st.Reject(t, BodyEquals("foo")(res, nil), nil)
	st.Reject(t, BodyEquals("")(res, nil), nil)

	body = ioutil.NopCloser(bytes.NewBufferString("hello world\n"))
	res = &http.Response{Body: body}
	st.Expect(t, BodyEquals("hello world")(res, nil), nil)
	st.Reject(t, BodyEquals("hello world\n")(res, nil), nil)
	st.Reject(t, BodyEquals("hello")(res, nil), nil)
	st.Reject(t, BodyEquals("foo")(res, nil), nil)
	st.Reject(t, BodyEquals("")(res, nil), nil)
}

func TestBodySnap(t *testing.T) {
	os.RemoveAll(Directory)
	t.Run("body-snap", func(t *testing.T) {
		name := "find-me-snap-test"
		b, _ := json.MarshalIndent(map[string]interface{}{
			"bool":   true,
			"number": 46,
			"float":  46.1,
			"string": "golang",
		}, "", "  ")
		url, _ := url.Parse("http://nickrandall.com")
		req := &http.Request{URL: url}

		ignoredFields := map[string]FieldFunc{}

		body := ioutil.NopCloser(bytes.NewBuffer(b))
		res := &http.Response{Body: body}
		st.Expect(t, BodySnap(name, ignoredFields)(res, req), nil)

		body = ioutil.NopCloser(bytes.NewBuffer(b))
		res = &http.Response{Body: body}
		st.Expect(t, BodySnap(name, ignoredFields)(res, req), nil)

		// TODO: add tests
		// add test to make sure file is created and json is what is expected
		file := Directory + "/" + name + "-body.snap"
		_, err := os.Stat(file)
		st.Expect(t, os.IsNotExist(err), false)
		fileContents, _ := ioutil.ReadFile(file)
		st.Expect(t, string(b), string(fileContents))

		b, _ = json.Marshal(map[string]interface{}{
			"bool":   true,
			"number": 46,
			"float":  46.1,
		})
		body = ioutil.NopCloser(bytes.NewBuffer(b))
		res = &http.Response{Body: body}
		st.Reject(t, BodySnap(name, ignoredFields)(res, req), nil)
	})

	fmt.Println("[Baloo]: Cleaned up snapshots")
	os.RemoveAll(Directory)
}
