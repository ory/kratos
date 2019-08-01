package password

import "net/url"

func tidyForm(vv url.Values) url.Values {
	for _, k := range []string{"password", "csrf_token", "request"} {
		vv.Del(k)
	}

	return vv
}
