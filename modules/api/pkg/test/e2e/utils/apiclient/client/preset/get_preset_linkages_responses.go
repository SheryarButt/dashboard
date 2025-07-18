// Code generated by go-swagger; DO NOT EDIT.

package preset

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"k8c.io/dashboard/v2/pkg/test/e2e/utils/apiclient/models"
)

// GetPresetLinkagesReader is a Reader for the GetPresetLinkages structure.
type GetPresetLinkagesReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetPresetLinkagesReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetPresetLinkagesOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 401:
		result := NewGetPresetLinkagesUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewGetPresetLinkagesForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewGetPresetLinkagesNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		result := NewGetPresetLinkagesDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetPresetLinkagesOK creates a GetPresetLinkagesOK with default headers values
func NewGetPresetLinkagesOK() *GetPresetLinkagesOK {
	return &GetPresetLinkagesOK{}
}

/*
GetPresetLinkagesOK describes a response with status code 200, with default header values.

PresetLinkages
*/
type GetPresetLinkagesOK struct {
	Payload *models.PresetLinkages
}

// IsSuccess returns true when this get preset linkages o k response has a 2xx status code
func (o *GetPresetLinkagesOK) IsSuccess() bool {
	return true
}

// IsRedirect returns true when this get preset linkages o k response has a 3xx status code
func (o *GetPresetLinkagesOK) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get preset linkages o k response has a 4xx status code
func (o *GetPresetLinkagesOK) IsClientError() bool {
	return false
}

// IsServerError returns true when this get preset linkages o k response has a 5xx status code
func (o *GetPresetLinkagesOK) IsServerError() bool {
	return false
}

// IsCode returns true when this get preset linkages o k response a status code equal to that given
func (o *GetPresetLinkagesOK) IsCode(code int) bool {
	return code == 200
}

func (o *GetPresetLinkagesOK) Error() string {
	return fmt.Sprintf("[GET /api/v2/presets/{preset_name}/linkages][%d] getPresetLinkagesOK  %+v", 200, o.Payload)
}

func (o *GetPresetLinkagesOK) String() string {
	return fmt.Sprintf("[GET /api/v2/presets/{preset_name}/linkages][%d] getPresetLinkagesOK  %+v", 200, o.Payload)
}

func (o *GetPresetLinkagesOK) GetPayload() *models.PresetLinkages {
	return o.Payload
}

func (o *GetPresetLinkagesOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.PresetLinkages)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetPresetLinkagesUnauthorized creates a GetPresetLinkagesUnauthorized with default headers values
func NewGetPresetLinkagesUnauthorized() *GetPresetLinkagesUnauthorized {
	return &GetPresetLinkagesUnauthorized{}
}

/*
GetPresetLinkagesUnauthorized describes a response with status code 401, with default header values.

EmptyResponse is a empty response
*/
type GetPresetLinkagesUnauthorized struct {
}

// IsSuccess returns true when this get preset linkages unauthorized response has a 2xx status code
func (o *GetPresetLinkagesUnauthorized) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get preset linkages unauthorized response has a 3xx status code
func (o *GetPresetLinkagesUnauthorized) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get preset linkages unauthorized response has a 4xx status code
func (o *GetPresetLinkagesUnauthorized) IsClientError() bool {
	return true
}

// IsServerError returns true when this get preset linkages unauthorized response has a 5xx status code
func (o *GetPresetLinkagesUnauthorized) IsServerError() bool {
	return false
}

// IsCode returns true when this get preset linkages unauthorized response a status code equal to that given
func (o *GetPresetLinkagesUnauthorized) IsCode(code int) bool {
	return code == 401
}

func (o *GetPresetLinkagesUnauthorized) Error() string {
	return fmt.Sprintf("[GET /api/v2/presets/{preset_name}/linkages][%d] getPresetLinkagesUnauthorized ", 401)
}

func (o *GetPresetLinkagesUnauthorized) String() string {
	return fmt.Sprintf("[GET /api/v2/presets/{preset_name}/linkages][%d] getPresetLinkagesUnauthorized ", 401)
}

func (o *GetPresetLinkagesUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewGetPresetLinkagesForbidden creates a GetPresetLinkagesForbidden with default headers values
func NewGetPresetLinkagesForbidden() *GetPresetLinkagesForbidden {
	return &GetPresetLinkagesForbidden{}
}

/*
GetPresetLinkagesForbidden describes a response with status code 403, with default header values.

EmptyResponse is a empty response
*/
type GetPresetLinkagesForbidden struct {
}

// IsSuccess returns true when this get preset linkages forbidden response has a 2xx status code
func (o *GetPresetLinkagesForbidden) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get preset linkages forbidden response has a 3xx status code
func (o *GetPresetLinkagesForbidden) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get preset linkages forbidden response has a 4xx status code
func (o *GetPresetLinkagesForbidden) IsClientError() bool {
	return true
}

// IsServerError returns true when this get preset linkages forbidden response has a 5xx status code
func (o *GetPresetLinkagesForbidden) IsServerError() bool {
	return false
}

// IsCode returns true when this get preset linkages forbidden response a status code equal to that given
func (o *GetPresetLinkagesForbidden) IsCode(code int) bool {
	return code == 403
}

func (o *GetPresetLinkagesForbidden) Error() string {
	return fmt.Sprintf("[GET /api/v2/presets/{preset_name}/linkages][%d] getPresetLinkagesForbidden ", 403)
}

func (o *GetPresetLinkagesForbidden) String() string {
	return fmt.Sprintf("[GET /api/v2/presets/{preset_name}/linkages][%d] getPresetLinkagesForbidden ", 403)
}

func (o *GetPresetLinkagesForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewGetPresetLinkagesNotFound creates a GetPresetLinkagesNotFound with default headers values
func NewGetPresetLinkagesNotFound() *GetPresetLinkagesNotFound {
	return &GetPresetLinkagesNotFound{}
}

/*
GetPresetLinkagesNotFound describes a response with status code 404, with default header values.

EmptyResponse is a empty response
*/
type GetPresetLinkagesNotFound struct {
}

// IsSuccess returns true when this get preset linkages not found response has a 2xx status code
func (o *GetPresetLinkagesNotFound) IsSuccess() bool {
	return false
}

// IsRedirect returns true when this get preset linkages not found response has a 3xx status code
func (o *GetPresetLinkagesNotFound) IsRedirect() bool {
	return false
}

// IsClientError returns true when this get preset linkages not found response has a 4xx status code
func (o *GetPresetLinkagesNotFound) IsClientError() bool {
	return true
}

// IsServerError returns true when this get preset linkages not found response has a 5xx status code
func (o *GetPresetLinkagesNotFound) IsServerError() bool {
	return false
}

// IsCode returns true when this get preset linkages not found response a status code equal to that given
func (o *GetPresetLinkagesNotFound) IsCode(code int) bool {
	return code == 404
}

func (o *GetPresetLinkagesNotFound) Error() string {
	return fmt.Sprintf("[GET /api/v2/presets/{preset_name}/linkages][%d] getPresetLinkagesNotFound ", 404)
}

func (o *GetPresetLinkagesNotFound) String() string {
	return fmt.Sprintf("[GET /api/v2/presets/{preset_name}/linkages][%d] getPresetLinkagesNotFound ", 404)
}

func (o *GetPresetLinkagesNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewGetPresetLinkagesDefault creates a GetPresetLinkagesDefault with default headers values
func NewGetPresetLinkagesDefault(code int) *GetPresetLinkagesDefault {
	return &GetPresetLinkagesDefault{
		_statusCode: code,
	}
}

/*
GetPresetLinkagesDefault describes a response with status code -1, with default header values.

errorResponse
*/
type GetPresetLinkagesDefault struct {
	_statusCode int

	Payload *models.ErrorResponse
}

// Code gets the status code for the get preset linkages default response
func (o *GetPresetLinkagesDefault) Code() int {
	return o._statusCode
}

// IsSuccess returns true when this get preset linkages default response has a 2xx status code
func (o *GetPresetLinkagesDefault) IsSuccess() bool {
	return o._statusCode/100 == 2
}

// IsRedirect returns true when this get preset linkages default response has a 3xx status code
func (o *GetPresetLinkagesDefault) IsRedirect() bool {
	return o._statusCode/100 == 3
}

// IsClientError returns true when this get preset linkages default response has a 4xx status code
func (o *GetPresetLinkagesDefault) IsClientError() bool {
	return o._statusCode/100 == 4
}

// IsServerError returns true when this get preset linkages default response has a 5xx status code
func (o *GetPresetLinkagesDefault) IsServerError() bool {
	return o._statusCode/100 == 5
}

// IsCode returns true when this get preset linkages default response a status code equal to that given
func (o *GetPresetLinkagesDefault) IsCode(code int) bool {
	return o._statusCode == code
}

func (o *GetPresetLinkagesDefault) Error() string {
	return fmt.Sprintf("[GET /api/v2/presets/{preset_name}/linkages][%d] getPresetLinkages default  %+v", o._statusCode, o.Payload)
}

func (o *GetPresetLinkagesDefault) String() string {
	return fmt.Sprintf("[GET /api/v2/presets/{preset_name}/linkages][%d] getPresetLinkages default  %+v", o._statusCode, o.Payload)
}

func (o *GetPresetLinkagesDefault) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *GetPresetLinkagesDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
