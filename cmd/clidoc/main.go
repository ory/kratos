package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/ory/kratos/cmd"
	"github.com/ory/kratos/text"

	"github.com/ory/x/clidoc"
)

var aSecondAgo = time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC).Add(-time.Second)
var inAMinute = time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC).Add(time.Minute)

var messages map[string]*text.Message

func init() {
	text.Now = func() time.Time {
		return inAMinute
	}
	text.Until = func(t time.Time) time.Duration {
		return time.Second
	}

	messages = map[string]*text.Message{
		"NewInfoNodeLabelVerifyOTP":                  text.NewInfoNodeLabelVerifyOTP(),
		"NewInfoNodeInputPassword":                   text.NewInfoNodeInputPassword(),
		"NewInfoNodeLabelGenerated":                  text.NewInfoNodeLabelGenerated("{title}"),
		"NewInfoNodeLabelSave":                       text.NewInfoNodeLabelSave(),
		"NewInfoNodeLabelSubmit":                     text.NewInfoNodeLabelSubmit(),
		"NewInfoNodeLabelID":                         text.NewInfoNodeLabelID(),
		"NewErrorValidationSettingsFlowExpired":      text.NewErrorValidationSettingsFlowExpired(time.Second),
		"NewInfoSelfServiceSettingsTOTPQRCode":       text.NewInfoSelfServiceSettingsTOTPQRCode(),
		"NewInfoSelfServiceSettingsTOTPSecret":       text.NewInfoSelfServiceSettingsTOTPSecret("{secret}"),
		"NewInfoSelfServiceSettingsTOTPSecretLabel":  text.NewInfoSelfServiceSettingsTOTPSecretLabel(),
		"NewInfoSelfServiceSettingsUpdateSuccess":    text.NewInfoSelfServiceSettingsUpdateSuccess(),
		"NewInfoSelfServiceSettingsUpdateUnlinkTOTP": text.NewInfoSelfServiceSettingsUpdateUnlinkTOTP(),
		"NewInfoSelfServiceSettingsRevealLookup":     text.NewInfoSelfServiceSettingsRevealLookup(),
		"NewInfoSelfServiceSettingsRegenerateLookup": text.NewInfoSelfServiceSettingsRegenerateLookup(),
		"NewInfoSelfServiceSettingsDisableLookup":    text.NewInfoSelfServiceSettingsDisableLookup(),
		"NewInfoSelfServiceSettingsLookupConfirm":    text.NewInfoSelfServiceSettingsLookupConfirm(),
		"NewInfoSelfServiceSettingsLookupSecretList": text.NewInfoSelfServiceSettingsLookupSecretList([]string{"{code-1}", "{code-2}"}, []interface{}{
			text.NewInfoSelfServiceSettingsLookupSecret("{code}"),
			text.NewInfoSelfServiceSettingsLookupSecretUsed(aSecondAgo),
		}),
		"NewInfoSelfServiceSettingsLookupSecret":                  text.NewInfoSelfServiceSettingsLookupSecret("{secret}"),
		"NewInfoSelfServiceSettingsLookupSecretUsed":              text.NewInfoSelfServiceSettingsLookupSecretUsed(aSecondAgo),
		"NewInfoSelfServiceSettingsLookupSecretsLabel":            text.NewInfoSelfServiceSettingsLookupSecretsLabel(),
		"NewInfoSelfServiceSettingsUpdateLinkOIDC":                text.NewInfoSelfServiceSettingsUpdateLinkOIDC("{provider}"),
		"NewInfoSelfServiceSettingsUpdateUnlinkOIDC":              text.NewInfoSelfServiceSettingsUpdateUnlinkOIDC("{provider}"),
		"NewInfoSelfServiceRegisterWebAuthn":                      text.NewInfoSelfServiceRegisterWebAuthn(),
		"NewInfoSelfServiceRegisterWebAuthnDisplayName":           text.NewInfoSelfServiceRegisterWebAuthnDisplayName(),
		"NewInfoSelfServiceRemoveWebAuthn":                        text.NewInfoSelfServiceRemoveWebAuthn("{name}", aSecondAgo),
		"NewErrorValidationVerificationFlowExpired":               text.NewErrorValidationVerificationFlowExpired(-time.Second),
		"NewInfoSelfServiceVerificationSuccessful":                text.NewInfoSelfServiceVerificationSuccessful(),
		"NewVerificationEmailSent":                                text.NewVerificationEmailSent(),
		"NewErrorValidationVerificationTokenInvalidOrAlreadyUsed": text.NewErrorValidationVerificationTokenInvalidOrAlreadyUsed(),
		"NewErrorValidationVerificationRetrySuccess":              text.NewErrorValidationVerificationRetrySuccess(),
		"NewErrorValidationVerificationStateFailure":              text.NewErrorValidationVerificationStateFailure(),
		"NewErrorSystemGeneric":                                   text.NewErrorSystemGeneric("{reason}"),
		"NewValidationErrorGeneric":                               text.NewValidationErrorGeneric("{reason}"),
		"NewValidationErrorRequired":                              text.NewValidationErrorRequired("{field}"),
		"NewErrorValidationMinLength":                             text.NewErrorValidationMinLength(1, 2),
		"NewErrorValidationInvalidFormat":                         text.NewErrorValidationInvalidFormat("{format}", "{value}"),
		"NewErrorValidationPasswordPolicyViolation":               text.NewErrorValidationPasswordPolicyViolation("{reason}"),
		"NewErrorValidationInvalidCredentials":                    text.NewErrorValidationInvalidCredentials(),
		"NewErrorValidationDuplicateCredentials":                  text.NewErrorValidationDuplicateCredentials(),
		"NewErrorValidationTOTPVerifierWrong":                     text.NewErrorValidationTOTPVerifierWrong(),
		"NewErrorValidationLookupAlreadyUsed":                     text.NewErrorValidationLookupAlreadyUsed(),
		"NewErrorValidationLookupInvalid":                         text.NewErrorValidationLookupInvalid(),
		"NewErrorValidationIdentifierMissing":                     text.NewErrorValidationIdentifierMissing(),
		"NewErrorValidationAddressNotVerified":                    text.NewErrorValidationAddressNotVerified(),
		"NewErrorValidationNoTOTPDevice":                          text.NewErrorValidationNoTOTPDevice(),
		"NewErrorValidationNoLookup":                              text.NewErrorValidationNoLookup(),
		"NewErrorValidationNoWebAuthnDevice":                      text.NewErrorValidationNoWebAuthnDevice(),
		"NewInfoLoginReAuth":                                      text.NewInfoLoginReAuth(),
		"NewInfoLoginMFA":                                         text.NewInfoLoginMFA(),
		"NewInfoLoginTOTPLabel":                                   text.NewInfoLoginTOTPLabel(),
		"NewInfoLoginLookupLabel":                                 text.NewInfoLoginLookupLabel(),
		"NewInfoLogin":                                            text.NewInfoLogin(),
		"NewInfoLoginTOTP":                                        text.NewInfoLoginTOTP(),
		"NewInfoLoginLookup":                                      text.NewInfoLoginLookup(),
		"NewInfoLoginVerify":                                      text.NewInfoLoginVerify(),
		"NewInfoLoginWith":                                        text.NewInfoLoginWith("{provider}"),
		"NewErrorValidationLoginFlowExpired":                      text.NewErrorValidationLoginFlowExpired(time.Second),
		"NewErrorValidationLoginNoStrategyFound":                  text.NewErrorValidationLoginNoStrategyFound(),
		"NewErrorValidationRegistrationNoStrategyFound":           text.NewErrorValidationRegistrationNoStrategyFound(),
		"NewErrorValidationSettingsNoStrategyFound":               text.NewErrorValidationSettingsNoStrategyFound(),
		"NewErrorValidationRecoveryNoStrategyFound":               text.NewErrorValidationRecoveryNoStrategyFound(),
		"NewErrorValidationVerificationNoStrategyFound":           text.NewErrorValidationVerificationNoStrategyFound(),
		"NewInfoSelfServiceLoginWebAuthn":                         text.NewInfoSelfServiceLoginWebAuthn(),
		"NewInfoRegistration":                                     text.NewInfoRegistration(),
		"NewInfoRegistrationWith":                                 text.NewInfoRegistrationWith("{provider}"),
		"NewInfoRegistrationContinue":                             text.NewInfoRegistrationContinue(),
		"NewErrorValidationRegistrationFlowExpired":               text.NewErrorValidationRegistrationFlowExpired(time.Second),
		"NewErrorValidationRecoveryFlowExpired":                   text.NewErrorValidationRecoveryFlowExpired(time.Second),
		"NewRecoverySuccessful":                                   text.NewRecoverySuccessful(inAMinute),
		"NewRecoveryEmailSent":                                    text.NewRecoveryEmailSent(),
		"NewErrorValidationRecoveryTokenInvalidOrAlreadyUsed":     text.NewErrorValidationRecoveryTokenInvalidOrAlreadyUsed(),
		"NewErrorValidationRecoveryRetrySuccess":                  text.NewErrorValidationRecoveryRetrySuccess(),
		"NewErrorValidationRecoveryStateFailure":                  text.NewErrorValidationRecoveryStateFailure(),
		"NewInfoNodeInputEmail":                                   text.NewInfoNodeInputEmail(),
	}
}

func main() {
	if err := clidoc.Generate(cmd.NewRootCmd(), os.Args[1:]); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to generate CLI docs: %+v", err)
		os.Exit(1)
	}

	if err := validateAllMessages(filepath.Join(os.Args[1], "text")); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to validate messages: %+v", err)
		os.Exit(1)
	}

	if err := writeMessages(filepath.Join(os.Args[1], "docs/docs/concepts/ui-user-interface.mdx")); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to generate message table: %+v", err)
		os.Exit(1)
	}

	fmt.Println("All files have been generated and updated.")
}

func codeEncode(in interface{}) string {
	out, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to encode to JSON: %+v", err)
		os.Exit(1)
	}

	return string(out)
}

func writeMessages(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var toSort []*text.Message
	for _, m := range messages {
		toSort = append(toSort, m)
	}

	sort.Slice(toSort, func(i, j int) bool {
		if toSort[i].ID == toSort[j].ID {
			return toSort[i].Text < toSort[j].Text
		}
		return toSort[i].ID < toSort[j].ID
	})

	var w bytes.Buffer
	for _, m := range toSort {
		w.WriteString(fmt.Sprintf(`###### %s (%d)

%s

`, m.Text, m.ID, "```json\n"+codeEncode(m)+"\n```"))
	}

	r := regexp.MustCompile(`(?s)<!-- START MESSAGE TABLE -->(.*?)<!-- END MESSAGE TABLE -->`)
	result := r.ReplaceAllString(string(content), "<!-- START MESSAGE TABLE -->\n"+w.String()+"\n<!-- END MESSAGE TABLE -->")

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	if _, err := f.WriteString(result); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

func validateAllMessages(path string) error {
	set := token.NewFileSet()
	packs, err := parser.ParseDir(set, path, nil, 0)
	if err != nil {
		return errors.Wrapf(err, "unable to parse text directory")
	}

	for _, pack := range packs {
		for _, f := range pack.Files {
			for _, d := range f.Decls {
				if fn, isFn := d.(*ast.FuncDecl); isFn {
					if name := fn.Name.String(); fn.Name.IsExported() && strings.HasPrefix(name, "New") {
						if _, ok := messages[name]; !ok {
							return errors.Errorf("expected to find message %s in the list for the documentation generation but could not", name)
						}
					}
				}
			}
		}
	}

	return nil
}
