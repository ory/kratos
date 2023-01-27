// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package password

import "net/url"

func tidyForm(vv url.Values) url.Values {
	for _, k := range []string{"password", "csrf_token", "flow"} {
		vv.Del(k)
	}

	return vv
}
