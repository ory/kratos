// Code generated by go-swagger; DO NOT EDIT.

package public

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/ory/kratos-client-go/models"
)

// GetSelfServiceLoginFlowReader is a Reader for the GetSelfServiceLoginFlow structure.
type GetSelfServiceLoginFlowReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetSelfServiceLoginFlowReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetSelfServiceLoginFlowOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 403:
		result := NewGetSelfServiceLoginFlowForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetSelfServiceLoginFlowNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 410:
		result := NewGetSelfServiceLoginFlowGone()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetSelfServiceLoginFlowInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGetSelfServiceLoginFlowOK creates a GetSelfServiceLoginFlowOK with default headers values
func NewGetSelfServiceLoginFlowOK() *GetSelfServiceLoginFlowOK {
	return &GetSelfServiceLoginFlowOK{}
}

/*GetSelfServiceLoginFlowOK handles this case with default header values.

loginFlow
*/
type GetSelfServiceLoginFlowOK struct {
	Payload *models.LoginFlow
}

func (o *GetSelfServiceLoginFlowOK) Error() string {
	return fmt.Sprintf("[GET /self-service/login/flows][%d] getSelfServiceLoginFlowOK  %+v", 200, o.Payload)
}

func (o *GetSelfServiceLoginFlowOK) GetPayload() *models.LoginFlow {
	return o.Payload
}

func (o *GetSelfServiceLoginFlowOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.LoginFlow)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetSelfServiceLoginFlowForbidden creates a GetSelfServiceLoginFlowForbidden with default headers values
func NewGetSelfServiceLoginFlowForbidden() *GetSelfServiceLoginFlowForbidden {
	return &GetSelfServiceLoginFlowForbidden{}
}

/*GetSelfServiceLoginFlowForbidden handles this case with default header values.

genericError
*/
type GetSelfServiceLoginFlowForbidden struct {
	Payload *models.GenericError
}

func (o *GetSelfServiceLoginFlowForbidden) Error() string {
	return fmt.Sprintf("[GET /self-service/login/flows][%d] getSelfServiceLoginFlowForbidden  %+v", 403, o.Payload)
}

func (o *GetSelfServiceLoginFlowForbidden) GetPayload() *models.GenericError {
	return o.Payload
}

func (o *GetSelfServiceLoginFlowForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.GenericError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetSelfServiceLoginFlowNotFound creates a GetSelfServiceLoginFlowNotFound with default headers values
func NewGetSelfServiceLoginFlowNotFound() *GetSelfServiceLoginFlowNotFound {
	return &GetSelfServiceLoginFlowNotFound{}
}

/*GetSelfServiceLoginFlowNotFound handles this case with default header values.

genericError
*/
type GetSelfServiceLoginFlowNotFound struct {
	Payload *models.GenericError
}

func (o *GetSelfServiceLoginFlowNotFound) Error() string {
	return fmt.Sprintf("[GET /self-service/login/flows][%d] getSelfServiceLoginFlowNotFound  %+v", 404, o.Payload)
}

func (o *GetSelfServiceLoginFlowNotFound) GetPayload() *models.GenericError {
	return o.Payload
}

func (o *GetSelfServiceLoginFlowNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.GenericError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetSelfServiceLoginFlowGone creates a GetSelfServiceLoginFlowGone with default headers values
func NewGetSelfServiceLoginFlowGone() *GetSelfServiceLoginFlowGone {
	return &GetSelfServiceLoginFlowGone{}
}

/*GetSelfServiceLoginFlowGone handles this case with default header values.

genericError
*/
type GetSelfServiceLoginFlowGone struct {
	Payload *models.GenericError
}

func (o *GetSelfServiceLoginFlowGone) Error() string {
	return fmt.Sprintf("[GET /self-service/login/flows][%d] getSelfServiceLoginFlowGone  %+v", 410, o.Payload)
}

func (o *GetSelfServiceLoginFlowGone) GetPayload() *models.GenericError {
	return o.Payload
}

func (o *GetSelfServiceLoginFlowGone) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.GenericError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetSelfServiceLoginFlowInternalServerError creates a GetSelfServiceLoginFlowInternalServerError with default headers values
func NewGetSelfServiceLoginFlowInternalServerError() *GetSelfServiceLoginFlowInternalServerError {
	return &GetSelfServiceLoginFlowInternalServerError{}
}

/*GetSelfServiceLoginFlowInternalServerError handles this case with default header values.

genericError
*/
type GetSelfServiceLoginFlowInternalServerError struct {
	Payload *models.GenericError
}

func (o *GetSelfServiceLoginFlowInternalServerError) Error() string {
	return fmt.Sprintf("[GET /self-service/login/flows][%d] getSelfServiceLoginFlowInternalServerError  %+v", 500, o.Payload)
}

func (o *GetSelfServiceLoginFlowInternalServerError) GetPayload() *models.GenericError {
	return o.Payload
}

func (o *GetSelfServiceLoginFlowInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.GenericError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
