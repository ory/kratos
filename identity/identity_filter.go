package identity

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/tidwall/gjson"
)

func Filter(values url.Values) ([]Identity, error) {
	var identities []Identity
	for _, identity := range identities {
		add := true
		for field, contents := range values {
			if field == "page" || field == "per_page" {
				continue
			}
			fmt.Printf("Filter in 2nd loop\n")
			raw, err := json.Marshal(identity)
			if err != nil {
				return []Identity{}, err
			}

			fmt.Printf("raw=%+v\n", string(raw))
			field, innerField := extractFieldAndInnerFields(field)
			res := gjson.GetBytes(raw, field)
			switch res.Type {
			case gjson.String:
				fmt.Printf("check in string=\n")
				if !IsStringInSlice(contents, res.String()) {
					add = false
					break
				}
			case gjson.JSON:
				present := false
				if innerField != "" {
					ra := res.Array()
					for _, r := range ra {
						if IsStringInSlice(contents, r.Get(innerField).String()) {
							present = true
							break
						}
					}
					if len(ra) == 0 {
						if IsStringInSlice(contents, res.Get(innerField).String()) {
							present = true
							break
						}
					}
				}
				add = add && present
			default:
				add = false
			}
		}
		if add {
			identities = append(identities, identity)
		}
	}
	return identities, nil
}

func IsStringInSlice(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func extractFieldAndInnerFields(field string) (string, string) {
	if !strings.Contains(field, ".") {
		return field, ""
	}

	dotIndex := strings.Index(field, ".")
	return field[:dotIndex], field[dotIndex+1:]
}
