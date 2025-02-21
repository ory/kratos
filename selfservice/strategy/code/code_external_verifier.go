package code

import (
	"context"
	"github.com/ory/kratos/courier"
	"github.com/ory/kratos/driver/config"
	"github.com/ory/kratos/request"
	"github.com/ory/kratos/x"
	"github.com/ory/x/jsonnetsecure"
	"github.com/ory/x/otelx"
	"github.com/pkg/errors"
	"net/http"
)

type (
	externalVerifierDependencies interface {
		config.Provider
		x.LoggingProvider
		x.TracingProvider
		x.HTTPClientProvider
		jsonnetsecure.VMProvider
	}

	ExternalVerifierProvider interface {
		ExternalVerifier() *ExternalVerifier
	}

	ExternalVerifier struct {
		deps externalVerifierDependencies
	}
)

func NewExternalVerifier(deps externalVerifierDependencies) *ExternalVerifier {
	return &ExternalVerifier{
		deps: deps,
	}
}

func (v *ExternalVerifier) VerificationStart(ctx context.Context, t courier.SMSTemplate) (err error) {
	ctx, span := v.deps.Tracer(ctx).Tracer().Start(ctx, "ExternalVerifier.VerificationStart")
	defer otelx.End(span, &err)

	builder, err := request.NewBuilder(ctx, v.deps.Config().SelfServiceCodeStrategy(ctx).ExternalSMSVerify.VerificationStartRequest, v.deps, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	req, err := builder.BuildRequest(ctx, t)
	if err != nil {
		return errors.WithStack(err)
	}
	req = req.WithContext(ctx)

	res, err := v.deps.HTTPClient(ctx).Do(req)
	if err != nil {
		return errors.WithStack(err)
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		phoneNumber, err := t.PhoneNumber()
		if err != nil {
			return errors.WithStack(err)
		}
		v.deps.Logger().
			WithField("to", phoneNumber).
			WithField("template_type", t.TemplateType).
			Debug("ExternalVerifier has started new verification.")
		return nil
	}

	body := x.MustReadAll(res.Body)
	if err := res.Body.Close(); err != nil {
		return err
	}

	err = errors.Errorf(
		"upstream server replied with status code %d and body %s",
		res.StatusCode,
		body,
	)
	return errors.WithStack(err)
}

func (v *ExternalVerifier) VerificationCheck(ctx context.Context, t courier.SMSTemplate) (err error) {
	ctx, span := v.deps.Tracer(ctx).Tracer().Start(ctx, "ExternalVerifier.VerificationCheck")
	defer otelx.End(span, &err)

	builder, err := request.NewBuilder(ctx, v.deps.Config().SelfServiceCodeStrategy(ctx).ExternalSMSVerify.VerificationCheckRequest, v.deps, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	req, err := builder.BuildRequest(ctx, t)
	if err != nil {
		return errors.WithStack(err)
	}
	req = req.WithContext(ctx)

	res, err := v.deps.HTTPClient(ctx).Do(req)
	if err != nil {
		return errors.WithStack(err)
	}

	if res.StatusCode >= 200 && res.StatusCode < 300 {
		phoneNumber, err := t.PhoneNumber()
		if err != nil {
			return errors.WithStack(err)
		}
		v.deps.Logger().
			WithField("to", phoneNumber).
			WithField("template_type", t.TemplateType).
			Debug("ExternalVerifier has checked a code.")
		return nil
	} else if res.StatusCode == http.StatusBadRequest {
		return errors.WithStack(ErrCodeNotFound)
	}

	body := x.MustReadAll(res.Body)
	if err := res.Body.Close(); err != nil {
		return err
	}

	err = errors.Errorf(
		"upstream server replied with status code %d and body %s",
		res.StatusCode,
		body,
	)
	return errors.WithStack(err)
}
