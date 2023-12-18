package anyjson

import (
	"encoding/json"
	testingx "github.com/octohelm/x/testing"
	"testing"
)

var data = []byte(`
{  
    "employees": [
		{  
			"name":      "octo",   
			"salary":     56000,   
			"married":    false  
		}  
	] 
}
`)

func TestUnmarshal(t *testing.T) {
	var obj Object

	err := json.Unmarshal(data, &obj)
	testingx.Expect(t, err, testingx.Be[error](nil))

	testingx.Expect(t, obj.Value(), testingx.Equal[any](Map{
		"employees": List{
			Map{
				"name":    "octo",
				"salary":  56000.0,
				"married": false,
			},
		},
	}))
}
