// Copyright © 2022 Ory Corp

package password

import "net/url"

func tidyForm(vv url.Values) url.Values {
	for _, k := range []string{"password", "csrf_token", "flow"} {
		vv.Del(k)
	}

	return vv
}
