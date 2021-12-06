package text

import (
	"encoding/json"
)

func context(ctx map[string]interface{}) []byte {
	if len(ctx) == 0 {
		return []byte("{}")
	}
	res, err := json.Marshal(ctx)
	if err != nil {
		panic(err)
	}
	return res
}
