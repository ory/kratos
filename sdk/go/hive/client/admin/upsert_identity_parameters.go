// Code generated by go-swagger; DO NOT EDIT.

package admin

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/ory/hive/sdk/go/hive/models"
)

// NewUpsertIdentityParams creates a new UpsertIdentityParams object
// with the default values initialized.
func NewUpsertIdentityParams() *UpsertIdentityParams {
	var ()
	return &UpsertIdentityParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewUpsertIdentityParamsWithTimeout creates a new UpsertIdentityParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewUpsertIdentityParamsWithTimeout(timeout time.Duration) *UpsertIdentityParams {
	var ()
	return &UpsertIdentityParams{

		timeout: timeout,
	}
}

// NewUpsertIdentityParamsWithContext creates a new UpsertIdentityParams object
// with the default values initialized, and the ability to set a context for a request
func NewUpsertIdentityParamsWithContext(ctx context.Context) *UpsertIdentityParams {
	var ()
	return &UpsertIdentityParams{

		Context: ctx,
	}
}

// NewUpsertIdentityParamsWithHTTPClient creates a new UpsertIdentityParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewUpsertIdentityParamsWithHTTPClient(client *http.Client) *UpsertIdentityParams {
	var ()
	return &UpsertIdentityParams{
		HTTPClient: client,
	}
}

/*UpsertIdentityParams contains all the parameters to send to the API endpoint
for the upsert identity operation typically these are written to a http.Request
*/
type UpsertIdentityParams struct {

	/*Body*/
	Body *models.Identity

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the upsert identity params
func (o *UpsertIdentityParams) WithTimeout(timeout time.Duration) *UpsertIdentityParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the upsert identity params
func (o *UpsertIdentityParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the upsert identity params
func (o *UpsertIdentityParams) WithContext(ctx context.Context) *UpsertIdentityParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the upsert identity params
func (o *UpsertIdentityParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the upsert identity params
func (o *UpsertIdentityParams) WithHTTPClient(client *http.Client) *UpsertIdentityParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the upsert identity params
func (o *UpsertIdentityParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the upsert identity params
func (o *UpsertIdentityParams) WithBody(body *models.Identity) *UpsertIdentityParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the upsert identity params
func (o *UpsertIdentityParams) SetBody(body *models.Identity) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *UpsertIdentityParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Body != nil {
		if err := r.SetBodyParam(o.Body); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
