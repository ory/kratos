// Copyright © 2026 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// applySberClaimsToMapperTraitsOutput подставляет в identity.traits строки из userinfo Сбера
// (рекурсивно по вложенным объектам), чтобы сохранить регистр и значения из ответа провайдера,
// в том числе после Jsonnet и при вложенных ключах в схеме traits.
func applySberClaimsToMapperTraitsOutput(providerID string, claims *Claims, evaluated string) (string, error) {
	if claims == nil || !isSberProviderID(providerID) {
		return evaluated, nil
	}

	traits := gjson.Get(evaluated, "identity.traits")
	if !traits.IsObject() {
		return evaluated, nil
	}

	patched, err := patchTraitObjectForSberClaims([]byte(traits.Raw), claims)
	if err != nil {
		return "", err
	}

	out, err := sjson.SetRawBytes([]byte(evaluated), "identity.traits", patched)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return string(out), nil
}

// applySberClaimsToIdentityTraitsBytes повторно применяет строки из userinfo к уже собранным traits
// (после merge с формой или с существующей идентичностью).
func applySberClaimsToIdentityTraitsBytes(providerID string, claims *Claims, traits []byte) ([]byte, error) {
	if claims == nil || !isSberProviderID(providerID) || len(traits) == 0 {
		return traits, nil
	}
	return patchTraitObjectForSberClaims(traits, claims)
}

func patchTraitObjectForSberClaims(raw []byte, claims *Claims) ([]byte, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return raw, nil
	}

	parsed := gjson.ParseBytes(raw)
	if !parsed.IsObject() {
		return raw, nil
	}

	out := []byte(parsed.Raw)
	var patchErr error

	parsed.ForEach(func(key, value gjson.Result) bool {
		k := key.String()
		lk := strings.ToLower(strings.TrimSpace(k))
		switch {
		case value.IsObject():
			nested, err := patchTraitObjectForSberClaims([]byte(value.Raw), claims)
			if err != nil {
				patchErr = err
				return false
			}
			if string(nested) != value.Raw {
				var err2 error
				out, err2 = sjson.SetRawBytes(out, k, nested)
				if err2 != nil {
					patchErr = err2
					return false
				}
			}
		case value.Type == gjson.String:
			rep, ok := claimForTraitKey(lk, claims)
			if !ok || strings.TrimSpace(rep) == "" {
				return true
			}
			if rep == value.String() {
				return true
			}
			var err2 error
			out, err2 = sjson.SetBytes(out, k, rep)
			if err2 != nil {
				patchErr = err2
				return false
			}
		case value.Type == gjson.Null:
			// до завершения регистрации фронт часто шлёт null для опциональных полей
			rep, ok := claimForTraitKey(lk, claims)
			if !ok || strings.TrimSpace(rep) == "" {
				return true
			}
			var err2 error
			out, err2 = sjson.SetBytes(out, k, rep)
			if err2 != nil {
				patchErr = err2
				return false
			}
		default:
			// массивы, числа, bool — не трогаем
		}
		return true
	})

	return out, patchErr
}

func claimForTraitKey(lk string, claims *Claims) (string, bool) {
	switch {
	case claims.GivenName != "" && isGivenNameTraitKey(lk):
		return claims.GivenName, true
	case claims.FamilyName != "" && isFamilyNameTraitKey(lk):
		return claims.FamilyName, true
	case claims.MiddleName != "" && isMiddleNameTraitKey(lk):
		return claims.MiddleName, true
	case claims.Email != "" && isEmailTraitKey(lk):
		return claims.Email, true
	case claims.Birthdate != "" && isBirthTraitKey(lk):
		return claims.Birthdate, true
	case isPhoneTraitKey(lk):
		n := normalizeRussianMobilePlus79(claims.PhoneNumber)
		if n == "" {
			return "", false
		}
		return n, true
	case claims.Picture != "" && isAvatarTraitKey(lk):
		return claims.Picture, true
	case claims.City != "" && lk == "city":
		return claims.City, true
	case claims.Address != "" && isAddressTraitKey(lk):
		return claims.Address, true
	case claims.School != "" && lk == "school":
		return claims.School, true
	case claims.University != "" && isUniversityTraitKey(lk):
		return claims.University, true
	default:
		return "", false
	}
}

func isGivenNameTraitKey(lk string) bool {
	switch lk {
	case "given_name", "first_name", "firstname", "name_first", "fname":
		return true
	default:
		return false
	}
}

func isMiddleNameTraitKey(lk string) bool {
	switch lk {
	case "middle_name", "patronymic", "middlename", "middle_name_patronymic", "second_name":
		return true
	default:
		return false
	}
}

func isFamilyNameTraitKey(lk string) bool {
	switch lk {
	case "family_name", "last_name", "surname", "lastname", "name_last", "lname":
		return true
	default:
		return false
	}
}

func isPhoneTraitKey(lk string) bool {
	switch lk {
	case "phone_number", "phone", "msisdn", "mobile", "tel", "telephone":
		return true
	default:
		return false
	}
}

func isAvatarTraitKey(lk string) bool {
	switch lk {
	case "avatar_url", "picture", "photo", "profile_image", "picture_url", "avatar":
		return true
	default:
		return false
	}
}

func isAddressTraitKey(lk string) bool {
	switch lk {
	case "address", "street_address", "addr", "postal_address":
		return true
	default:
		return false
	}
}

func isUniversityTraitKey(lk string) bool {
	switch lk {
	case "university", "uni", "college", "higher_education":
		return true
	default:
		return false
	}
}

func isEmailTraitKey(lk string) bool {
	switch lk {
	case "email", "e_mail", "mail":
		return true
	default:
		return false
	}
}

func isBirthTraitKey(lk string) bool {
	if lk == "dob" {
		return true
	}
	// birth_date, birthdate, date_of_birth и т.п.
	if strings.Contains(lk, "birth") {
		return true
	}
	return false
}
