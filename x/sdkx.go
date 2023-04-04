// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package x

func SDKError(err error) error {
	if err == nil {
		return nil
	}

	if err.Error() == "" {
		return nil
	}

	return err
}
