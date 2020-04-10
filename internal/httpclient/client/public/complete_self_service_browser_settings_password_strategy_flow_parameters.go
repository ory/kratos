// Code generated by go-swagger; DO NOT EDIT.

package public

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
)

// NewCompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams creates a new CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams object
// with the default values initialized.
func NewCompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams() *CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams {

	return &CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewCompleteSelfServiceBrowserSettingsPasswordStrategyFlowParamsWithTimeout creates a new CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewCompleteSelfServiceBrowserSettingsPasswordStrategyFlowParamsWithTimeout(timeout time.Duration) *CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams {

	return &CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams{

		timeout: timeout,
	}
}

// NewCompleteSelfServiceBrowserSettingsPasswordStrategyFlowParamsWithContext creates a new CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams object
// with the default values initialized, and the ability to set a context for a request
func NewCompleteSelfServiceBrowserSettingsPasswordStrategyFlowParamsWithContext(ctx context.Context) *CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams {

	return &CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams{

		Context: ctx,
	}
}

// NewCompleteSelfServiceBrowserSettingsPasswordStrategyFlowParamsWithHTTPClient creates a new CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewCompleteSelfServiceBrowserSettingsPasswordStrategyFlowParamsWithHTTPClient(client *http.Client) *CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams {

	return &CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams{
		HTTPClient: client,
	}
}

/*CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams contains all the parameters to send to the API endpoint
for the complete self service browser settings password strategy flow operation typically these are written to a http.Request
*/
type CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the complete self service browser settings password strategy flow params
func (o *CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams) WithTimeout(timeout time.Duration) *CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the complete self service browser settings password strategy flow params
func (o *CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the complete self service browser settings password strategy flow params
func (o *CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams) WithContext(ctx context.Context) *CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the complete self service browser settings password strategy flow params
func (o *CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the complete self service browser settings password strategy flow params
func (o *CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams) WithHTTPClient(client *http.Client) *CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the complete self service browser settings password strategy flow params
func (o *CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *CompleteSelfServiceBrowserSettingsPasswordStrategyFlowParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
