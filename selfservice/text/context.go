package text

import "github.com/tidwall/sjson"

func context(ctx map[string]interface{}) []byte {
	json := `{}`
	var err error
	for key, value := range ctx {
		json, err = sjson.Set(json, key, value)
		if err != nil {
			panic(err)
		}
	}
	return []byte(json)
}
