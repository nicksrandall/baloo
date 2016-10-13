package assert_snapshot

import (
	"testing"

	"github.com/nicksrandall/baloo"
	"github.com/nicksrandall/baloo/assert"
)

// test stores the HTTP testing client preconfigured
var test = baloo.New("http://httpbin.org")

func TestBalooSnapshot(t *testing.T) {
	test.Get("/user-agent").
		SetHeader("Foo", "Bar").
		Expect(t).
		Status(200).
		Type("json").
		Field("user-agent", assert.FieldIsString).
		BodySnap().
		Done()
}
