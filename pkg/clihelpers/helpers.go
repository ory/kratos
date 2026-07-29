// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package clihelpers

const (
	WarningJQIsComplicated = "We have to admit, this is not easy if you don't speak jq fluently. What about opening an issue and telling us what predefined selectors you want to have? https://github.com/ory/kratos/issues/new/choose"
)

// SDKError normalizes an error returned by the generated SDK. The SDK reports
// some non-failures as errors with an empty message; those are not real errors.
func SDKError(err error) error {
	if err == nil {
		return nil
	}

	if err.Error() == "" {
		return nil
	}

	return err
}
