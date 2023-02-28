// Code generated by go-swagger; DO NOT EDIT.

// This file is part of MinIO Operator
// Copyright (c) 2023 MinIO, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

package operator_api

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/minio/operator/models"
)

// TenantDeleteEncryptionNoContentCode is the HTTP code returned for type TenantDeleteEncryptionNoContent
const TenantDeleteEncryptionNoContentCode int = 204

/*
TenantDeleteEncryptionNoContent A successful response.

swagger:response tenantDeleteEncryptionNoContent
*/
type TenantDeleteEncryptionNoContent struct {
}

// NewTenantDeleteEncryptionNoContent creates TenantDeleteEncryptionNoContent with default headers values
func NewTenantDeleteEncryptionNoContent() *TenantDeleteEncryptionNoContent {

	return &TenantDeleteEncryptionNoContent{}
}

// WriteResponse to the client
func (o *TenantDeleteEncryptionNoContent) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(204)
}

/*
TenantDeleteEncryptionDefault Generic error response.

swagger:response tenantDeleteEncryptionDefault
*/
type TenantDeleteEncryptionDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewTenantDeleteEncryptionDefault creates TenantDeleteEncryptionDefault with default headers values
func NewTenantDeleteEncryptionDefault(code int) *TenantDeleteEncryptionDefault {
	if code <= 0 {
		code = 500
	}

	return &TenantDeleteEncryptionDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the tenant delete encryption default response
func (o *TenantDeleteEncryptionDefault) WithStatusCode(code int) *TenantDeleteEncryptionDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the tenant delete encryption default response
func (o *TenantDeleteEncryptionDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the tenant delete encryption default response
func (o *TenantDeleteEncryptionDefault) WithPayload(payload *models.Error) *TenantDeleteEncryptionDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the tenant delete encryption default response
func (o *TenantDeleteEncryptionDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *TenantDeleteEncryptionDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
