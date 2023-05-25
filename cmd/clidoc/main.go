// Copyright © 2023 Ory Corp
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
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
		"NewInfoNodeLabelVerificationCode":           text.NewInfoNodeLabelVerificationCode(),
		"NewInfoNodeLabelRecoveryCode":               text.NewInfoNodeLabelRecoveryCode(),
		"NewInfoNodeInputPassword":                   text.NewInfoNodeInputPassword(),
		"NewInfoNodeLabelGenerated":                  text.NewInfoNodeLabelGenerated("{title}"),
		"NewInfoNodeLabelSave":                       text.NewInfoNodeLabelSave(),
		"NewInfoNodeLabelSubmit":                     text.NewInfoNodeLabelSubmit(),
		"NewInfoNodeLabelID":                         text.NewInfoNodeLabelID(),
		"NewErrorValidationSettingsFlowExpired":      text.NewErrorValidationSettingsFlowExpired(aSecondAgo),
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
		"NewInfoSelfServiceRegisterWebAuthn":                      text.NewInfoSelfServiceSettingsRegisterWebAuthn(),
		"NewInfoSelfServiceRegisterWebAuthnDisplayName":           text.NewInfoSelfServiceRegisterWebAuthnDisplayName(),
		"NewInfoSelfServiceRemoveWebAuthn":                        text.NewInfoSelfServiceRemoveWebAuthn("{name}", aSecondAgo),
		"NewErrorValidationVerificationFlowExpired":               text.NewErrorValidationVerificationFlowExpired(aSecondAgo),
		"NewInfoSelfServiceVerificationSuccessful":                text.NewInfoSelfServiceVerificationSuccessful(),
		"NewVerificationEmailSent":                                text.NewVerificationEmailSent(),
		"NewVerificationEmailWithCodeSent":                        text.NewVerificationEmailWithCodeSent(),
		"NewErrorValidationVerificationTokenInvalidOrAlreadyUsed": text.NewErrorValidationVerificationTokenInvalidOrAlreadyUsed(),
		"NewErrorValidationVerificationRetrySuccess":              text.NewErrorValidationVerificationRetrySuccess(),
		"NewErrorValidationVerificationStateFailure":              text.NewErrorValidationVerificationStateFailure(),
		"NewErrorValidationVerificationCodeInvalidOrAlreadyUsed":  text.NewErrorValidationVerificationCodeInvalidOrAlreadyUsed(),
		"NewErrorSystemGeneric":                                   text.NewErrorSystemGeneric("{reason}"),
		"NewValidationErrorGeneric":                               text.NewValidationErrorGeneric("{reason}"),
		"NewValidationErrorRequired":                              text.NewValidationErrorRequired("{field}"),
		"NewErrorValidationMinLength":                             text.NewErrorValidationMinLength("length must be >= 5, but got 3"),
		"NewErrorValidationMaxLength":                             text.NewErrorValidationMaxLength("length must be <= 5, but got 6"),
		"NewErrorValidationInvalidFormat":                         text.NewErrorValidationInvalidFormat("does not match pattern \"^[a-z]*$\""),
		"NewErrorValidationMinimum":                               text.NewErrorValidationMinimum("must be >= 5 but found 3"),
		"NewErrorValidationExclusiveMinimum":                      text.NewErrorValidationExclusiveMinimum("must be > 5 but found 5"),
		"NewErrorValidationMaximum":                               text.NewErrorValidationMaximum("must be <= 5 but found 6"),
		"NewErrorValidationExclusiveMaximum":                      text.NewErrorValidationExclusiveMaximum("must be < 5 but found 5"),
		"NewErrorValidationMultipleOf":                            text.NewErrorValidationMultipleOf("7 not multipleOf 3"),
		"NewErrorValidationMaxItems":                              text.NewErrorValidationMaxItems("maximum 3 items allowed, but found 4 items"),
		"NewErrorValidationMinItems":                              text.NewErrorValidationMinItems("minimum 3 items allowed, but found 2 items"),
		"NewErrorValidationUniqueItems":                           text.NewErrorValidationUniqueItems("items at index 0 and 2 are equal"),
		"NewErrorValidationWrongType":                             text.NewErrorValidationWrongType("expected number, but got string"),
		"NewErrorValidationPasswordPolicyViolation":               text.NewErrorValidationPasswordPolicyViolation("{reason}"),
		"NewErrorValidationInvalidCredentials":                    text.NewErrorValidationInvalidCredentials(),
		"NewErrorValidationDuplicateCredentials":                  text.NewErrorValidationDuplicateCredentials(),
		"NewErrorValidationDuplicateCredentialsOnOIDCLink":        text.NewErrorValidationDuplicateCredentialsOnOIDCLink(),
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
		"NewErrorValidationLoginFlowExpired":                      text.NewErrorValidationLoginFlowExpired(aSecondAgo),
		"NewErrorValidationLoginNoStrategyFound":                  text.NewErrorValidationLoginNoStrategyFound(),
		"NewErrorValidationRegistrationNoStrategyFound":           text.NewErrorValidationRegistrationNoStrategyFound(),
		"NewErrorValidationSettingsNoStrategyFound":               text.NewErrorValidationSettingsNoStrategyFound(),
		"NewErrorValidationRecoveryNoStrategyFound":               text.NewErrorValidationRecoveryNoStrategyFound(),
		"NewErrorValidationVerificationNoStrategyFound":           text.NewErrorValidationVerificationNoStrategyFound(),
		"NewInfoSelfServiceLoginWebAuthn":                         text.NewInfoSelfServiceLoginWebAuthn(),
		"NewInfoRegistration":                                     text.NewInfoRegistration(),
		"NewInfoRegistrationWith":                                 text.NewInfoRegistrationWith("{provider}"),
		"NewInfoRegistrationContinue":                             text.NewInfoRegistrationContinue(),
		"NewErrorValidationRegistrationFlowExpired":               text.NewErrorValidationRegistrationFlowExpired(aSecondAgo),
		"NewErrorValidationRecoveryFlowExpired":                   text.NewErrorValidationRecoveryFlowExpired(aSecondAgo),
		"NewRecoverySuccessful":                                   text.NewRecoverySuccessful(inAMinute),
		"NewRecoveryEmailSent":                                    text.NewRecoveryEmailSent(),
		"NewRecoveryEmailWithCodeSent":                            text.NewRecoveryEmailWithCodeSent(),
		"NewErrorValidationRecoveryTokenInvalidOrAlreadyUsed":     text.NewErrorValidationRecoveryTokenInvalidOrAlreadyUsed(),
		"NewErrorValidationRecoveryCodeInvalidOrAlreadyUsed":      text.NewErrorValidationRecoveryCodeInvalidOrAlreadyUsed(),
		"NewErrorValidationRecoveryRetrySuccess":                  text.NewErrorValidationRecoveryRetrySuccess(),
		"NewErrorValidationRecoveryStateFailure":                  text.NewErrorValidationRecoveryStateFailure(),
		"NewInfoNodeInputEmail":                                   text.NewInfoNodeInputEmail(),
		"NewInfoNodeResendOTP":                                    text.NewInfoNodeResendOTP(),
		"NewInfoNodeLabelContinue":                                text.NewInfoNodeLabelContinue(),
		"NewInfoSelfServiceSettingsRegisterWebAuthn":              text.NewInfoSelfServiceSettingsRegisterWebAuthn(),
		"NewInfoLoginWebAuthnPasswordless":                        text.NewInfoLoginWebAuthnPasswordless(),
		"NewInfoSelfServiceRegistrationRegisterWebAuthn":          text.NewInfoSelfServiceRegistrationRegisterWebAuthn(),
		"NewInfoLoginPasswordlessWebAuthn":                        text.NewInfoLoginPasswordlessWebAuthn(),
		"NewInfoSelfServiceContinueLoginWebAuthn":                 text.NewInfoSelfServiceContinueLoginWebAuthn(),
		"NewInfoSelfServiceLoginContinue":                         text.NewInfoSelfServiceLoginContinue(),
		"NewErrorValidationSuchNoWebAuthnUser":                    text.NewErrorValidationSuchNoWebAuthnUser(),
	}
}

func main() {
	if err := clidoc.Generate(cmd.NewRootCmd(), []string{filepath.Join(os.Args[2], "cli")}); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to generate CLI docs: %+v", err)
		os.Exit(1)
	}

	if err := validateAllMessages(filepath.Join(os.Args[1], "text")); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to validate messages: %+v\n", err)
		os.Exit(1)
	}

	sortedMessages := sortMessages()

	if err := writeMessages(filepath.Join(os.Args[2], "concepts/ui-user-interface.mdx"), sortedMessages); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to generate message table: %+v\n", err)
		os.Exit(1)
	}

	if err := writeMessagesJson(filepath.Join(os.Args[2], "concepts/messages.json"), sortedMessages); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Unable to generate messages.json: %+v\n", err)
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

func sortMessages() []*text.Message {
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

	return toSort
}

func writeMessages(path string, sortedMessages []*text.Message) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var w bytes.Buffer
	for _, m := range sortedMessages {
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

func writeMessagesJson(path string, sortedMessages []*text.Message) error {
	result := codeEncode(sortedMessages)

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
	type message struct {
		ID, Name string
	}

	usedIDs := make([]message, 0, len(messages))
	set := token.NewFileSet()
	packs, err := parser.ParseDir(set, path, nil, 0)
	if err != nil {
		return errors.Wrapf(err, "unable to parse text directory")
	}
	info := &types.Info{
		Defs: make(map[*ast.Ident]types.Object),
	}
	var pack *ast.Package
	for _, p := range packs {
		if p.Name == "text" {
			pack = p
			break
		}
	}
	allFiles := make([]*ast.File, 0)
	for fn, f := range pack.Files {
		if strings.HasSuffix(fn, "/id.go") {
			allFiles = append(allFiles, f)
		}
	}
	_, err = (&types.Config{Importer: importer.Default()}).Check("text", set, allFiles, info)
	if err != nil {
		return err // type error
	}

	for _, f := range pack.Files {
		for _, d := range f.Decls {
			switch decl := d.(type) {
			case *ast.FuncDecl:
				if name := decl.Name.String(); decl.Name.IsExported() && strings.HasPrefix(name, "New") {
					if _, ok := messages[name]; !ok {
						return errors.Errorf("expected to find message %s in the list for the documentation generation but could not", name)
					}
				}
			case *ast.GenDecl:
				if decl.Tok == token.CONST {
					for _, spec := range decl.Specs {
						value := spec.(*ast.ValueSpec) // safe because decl.Tok is token.CONST
						if t, ok := value.Type.(*ast.Ident); ok {
							if t.Name == "ID" {
								for _, name := range value.Names {
									c := info.ObjectOf(name)
									if c == nil {
										return errors.Errorf("expected to find const %s in text/id.go", name.Name)
									}
									usedIDs = append(usedIDs, message{
										ID:   c.(*types.Const).Val().ExactString(),
										Name: name.Name,
									})
								}
							}
						}
					}
				}
			}
		}
	}

	sort.Slice(usedIDs, func(i, j int) bool {
		return usedIDs[i].ID < usedIDs[j].ID
	})
	for i := 1; i < len(usedIDs); i++ {
		if usedIDs[i].ID == usedIDs[i-1].ID {
			return errors.Errorf("message ID %s is used more than once: %s %s", usedIDs[i].ID, usedIDs[i].Name, usedIDs[i-1].Name)
		}
	}

	return nil
}
