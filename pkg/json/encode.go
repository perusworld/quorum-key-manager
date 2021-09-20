package json

import (
	"database/sql/driver"
	"encoding/json"
)

var (
	_ driver.Valuer = jsonArray{}
	_ driver.Valuer = jsonMap{}
)

type jsonArray []interface{}

func (a jsonArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type jsonMap map[string]interface{}

func (m jsonMap) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// RecursiveToJSON recursively convert all map[interface]interface{} to map[string]interface{}
// as Go refuses to convert map[interface{}]interface{} to JSON because JSON only support string keys
func RecursiveToJSON(v interface{}) (r interface{}) {
	switch v := v.(type) {
	case []interface{}:
		for i, e := range v {
			v[i] = RecursiveToJSON(e)
		}
		r = jsonArray(v)
	case map[interface{}]interface{}:
		newMap := make(map[string]interface{}, len(v))
		for k, e := range v {
			newMap[k.(string)] = RecursiveToJSON(e)
		}
		r = jsonMap(newMap)
	default:
		r = v
	}
	return
}
