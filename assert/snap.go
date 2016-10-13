package assert

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"

	gentleman "gopkg.in/h2non/gentleman.v1"

	diff "github.com/nicksrandall/gojsondiff"
	"github.com/nicksrandall/gojsondiff/formatter"
)

const Directory = "__snapshots__"

var Update bool = false

func init() {
	os.MkdirAll(Directory, 0777)
	flag.BoolVar(&Update, "update-all", false, "use this flag to update all failing snapshots")
	flag.Parse()
}

func Shot(key string, data interface{}) (success bool, diff string, err error) {
	file := Directory + "/" + key + ".snap"
	if _, err := os.Stat(file); !os.IsNotExist(err) && !Update {
		dataFromSnap, err := ioutil.ReadFile(file)
		if err != nil {
			return false, "", err
		}

		dataFromTest, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return false, "", err
		}

		if !reflect.DeepEqual(dataFromSnap, dataFromTest) {
			var diffString string
			var err error
			switch d := data.(type) {
			case []interface{}:
				var snap []interface{}
				json.Unmarshal(dataFromSnap, &snap)
				diffString, err = printArrayDiff(d, snap)
			case map[string]interface{}:
				var snap map[string]interface{}
				json.Unmarshal(dataFromSnap, &snap)
				diffString, err = printJsonDiff(d, snap)
			case nil:
				var nothing map[string]interface{}
				var snap map[string]interface{}
				json.Unmarshal(dataFromSnap, &snap)
				diffString, err = printJsonDiff(nothing, snap)
			default:
				err = fmt.Errorf("Unrecognized type for Shot: %T", d)
			}
			return false, diffString, err
		}

		return true, "", nil

	} else {
		// write json to file
		err := writeSnapshotToFile(key, data)
		if err != nil {
			return false, "", err
		}
		return true, "", nil
	}
}

func writeSnapshotToFile(key string, data interface{}) error {
	file := Directory + "/" + key + ".snap"
	// write json to file
	dataFromTest, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(file)
	os.MkdirAll(dir, 0777)
	err = ioutil.WriteFile(file, dataFromTest, 0777)
	if err != nil {
		return err
	}
	fmt.Printf("[Baloo]: Added snapshot for key: %s\n", key)
	return nil
}

func MakeNewSnapshot(key string, resp *gentleman.Response) error {
	var data interface{}
	resp.JSON(&data)
	return writeSnapshotToFile(key, data)
}

func printJsonDiff(a, b map[string]interface{}) (string, error) {
	differ := diff.New()
	d := differ.CompareObjects(a, b)

	formatter := formatter.NewAsciiFormatter(a)
	formatter.ShowArrayIndex = true
	diffString, err := formatter.Format(d)
	if err != nil {
		return "", err
	}
	return diffString, nil
}

func printArrayDiff(a, b []interface{}) (string, error) {
	differ := diff.New()
	d := differ.CompareArrays(a, b)

	formatter := formatter.NewAsciiFormatter(a)
	formatter.ShowArrayIndex = true
	diffString, err := formatter.Format(d)
	if err != nil {
		return "", err
	}
	return diffString, nil
}

type SnapError struct {
	s string
}

func (e SnapError) Error() string {
	return e.s
}

func NewSnapError(s string) error {
	return SnapError{s}
}
