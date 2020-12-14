// nolint
package questions

import (
	"context"
	"regexp"
	"strings"

	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/hash"
	"github.com/ory/kratos/identity"
	"github.com/ory/kratos/selfservice/form"
)

type (
	ManagementProvider interface {
		RecoveryManager() *Manager
	}
	managerDependencies interface {
		hash.HashProvider
	}
	Manager struct {
		c config.Provider
		d managerDependencies
	}
)

var collapseNonAlphanumeric = regexp.MustCompile(`[^a-zA-Z\d\s:]+`)

func normalizeAnswer(in string) []byte {
	return []byte(collapseNonAlphanumeric.ReplaceAllString(strings.ToLower(strings.TrimSpace(in)), " "))
}

func (m *Manager) SetSecurityFormFields(ctx context.Context, i *identity.Identity, prefix string, htmlf *form.HTMLForm) error {
	if len(prefix) > 0 {
		prefix = prefix + "."
	}

	// for _, question := range i.RecoverySecurityAnswers {
	// 	htmlf.SetField(form.Field{Name: prefix + question.Key, Type: "text", Required: true})
	// }
	return nil
}

// func (m *Manager) CompareSecurityQuestions(ctx context.Context, question identity.RecoverySecurityAnswer, answer string) error {
// 	return m.d.Hasher().Compare([]byte(answer), normalizeAnswer(question.Answer))
// }

func (m *Manager) HashSecurityQuestions(i *identity.Identity, answers map[string]string) error {
	// var result identity.RecoverySecurityAnswers
	//
	// for key, answer := range answers {
	// 	hashed, err := m.d.Hasher().Generate(normalizeAnswer(answer))
	// 	if err != nil {
	// 		return err
	// 	}
	//
	// 	result = append(result,
	// 		identity.RecoverySecurityAnswer{ID: x.NewUUID(), Key: key, Answer: string(hashed), IdentityID: i.ID})
	// }
	//
	// i.RecoverySecurityAnswers = result
	return nil
}

func (m *Manager) SetSecurityAnswers(ctx context.Context, i *identity.Identity, answers map[string]string, validationPrefix string) error {
	// // Validation
	// for _, question := range m.c.SelfServiceRecoverySecurityQuestions() {
	// 	var found bool
	// 	for id, answer := range answers {
	// 		if id == question.ID {
	// 			expected := 6
	// 			answer = normalizeAnswer(answer)
	//
	// 			if actual := utf8.RuneCountInString(answer); actual < expected {
	// 				return errors.WithStack(&jsonschema.ValidationError{
	// 					Message:     fmt.Sprintf("length must be >= %d, but got %s (%d) after normalization", expected, answer, actual),
	// 					InstancePtr: validationPrefix + id,
	// 				})
	// 			}
	//
	// 			found = true
	// 			i.RecoverySecurityAnswers = append(i.RecoverySecurityAnswers, identity.RecoverySecurityAnswer{
	// 				Key: id, Answer: answer,
	// 				IdentityID: i.ID,
	// 			})
	// 			break
	// 		}
	// 	}
	//
	// 	if !found {
	// 		return schema.NewRequiredError(validationPrefix, question.ID)
	// 	}
	// }

	return nil
}
