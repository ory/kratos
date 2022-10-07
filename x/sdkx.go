// Copyright © 2022 Ory Corp

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
