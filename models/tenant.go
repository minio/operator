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

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"strconv"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// Tenant tenant
//
// swagger:model tenant
type Tenant struct {

	// creation date
	CreationDate string `json:"creation_date,omitempty"`

	// current state
	CurrentState string `json:"currentState,omitempty"`

	// deletion date
	DeletionDate string `json:"deletion_date,omitempty"`

	// domains
	Domains *DomainsConfiguration `json:"domains,omitempty"`

	// encryption enabled
	EncryptionEnabled bool `json:"encryptionEnabled,omitempty"`

	// endpoints
	Endpoints *TenantEndpoints `json:"endpoints,omitempty"`

	// idp ad enabled
	IdpAdEnabled bool `json:"idpAdEnabled,omitempty"`

	// idp oidc enabled
	IdpOidcEnabled bool `json:"idpOidcEnabled,omitempty"`

	// image
	Image string `json:"image,omitempty"`

	// log enabled
	LogEnabled bool `json:"logEnabled,omitempty"`

	// minio TLS
	MinioTLS bool `json:"minioTLS,omitempty"`

	// name
	Name string `json:"name,omitempty"`

	// namespace
	Namespace string `json:"namespace,omitempty"`

	// pools
	Pools []*Pool `json:"pools"`

	// status
	Status *TenantStatus `json:"status,omitempty"`

	// subnet license
	SubnetLicense *License `json:"subnet_license,omitempty"`

	// tiers
	Tiers []*TenantTierElement `json:"tiers"`

	// total size
	TotalSize int64 `json:"total_size,omitempty"`
}

// Validate validates this tenant
func (m *Tenant) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateDomains(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateEndpoints(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validatePools(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateStatus(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateSubnetLicense(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateTiers(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Tenant) validateDomains(formats strfmt.Registry) error {
	if swag.IsZero(m.Domains) { // not required
		return nil
	}

	if m.Domains != nil {
		if err := m.Domains.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("domains")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("domains")
			}
			return err
		}
	}

	return nil
}

func (m *Tenant) validateEndpoints(formats strfmt.Registry) error {
	if swag.IsZero(m.Endpoints) { // not required
		return nil
	}

	if m.Endpoints != nil {
		if err := m.Endpoints.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("endpoints")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("endpoints")
			}
			return err
		}
	}

	return nil
}

func (m *Tenant) validatePools(formats strfmt.Registry) error {
	if swag.IsZero(m.Pools) { // not required
		return nil
	}

	for i := 0; i < len(m.Pools); i++ {
		if swag.IsZero(m.Pools[i]) { // not required
			continue
		}

		if m.Pools[i] != nil {
			if err := m.Pools[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("pools" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("pools" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

func (m *Tenant) validateStatus(formats strfmt.Registry) error {
	if swag.IsZero(m.Status) { // not required
		return nil
	}

	if m.Status != nil {
		if err := m.Status.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("status")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("status")
			}
			return err
		}
	}

	return nil
}

func (m *Tenant) validateSubnetLicense(formats strfmt.Registry) error {
	if swag.IsZero(m.SubnetLicense) { // not required
		return nil
	}

	if m.SubnetLicense != nil {
		if err := m.SubnetLicense.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("subnet_license")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("subnet_license")
			}
			return err
		}
	}

	return nil
}

func (m *Tenant) validateTiers(formats strfmt.Registry) error {
	if swag.IsZero(m.Tiers) { // not required
		return nil
	}

	for i := 0; i < len(m.Tiers); i++ {
		if swag.IsZero(m.Tiers[i]) { // not required
			continue
		}

		if m.Tiers[i] != nil {
			if err := m.Tiers[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("tiers" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("tiers" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// ContextValidate validate this tenant based on the context it is used
func (m *Tenant) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateDomains(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateEndpoints(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidatePools(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateStatus(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateSubnetLicense(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateTiers(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Tenant) contextValidateDomains(ctx context.Context, formats strfmt.Registry) error {

	if m.Domains != nil {
		if err := m.Domains.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("domains")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("domains")
			}
			return err
		}
	}

	return nil
}

func (m *Tenant) contextValidateEndpoints(ctx context.Context, formats strfmt.Registry) error {

	if m.Endpoints != nil {
		if err := m.Endpoints.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("endpoints")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("endpoints")
			}
			return err
		}
	}

	return nil
}

func (m *Tenant) contextValidatePools(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.Pools); i++ {

		if m.Pools[i] != nil {
			if err := m.Pools[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("pools" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("pools" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

func (m *Tenant) contextValidateStatus(ctx context.Context, formats strfmt.Registry) error {

	if m.Status != nil {
		if err := m.Status.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("status")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("status")
			}
			return err
		}
	}

	return nil
}

func (m *Tenant) contextValidateSubnetLicense(ctx context.Context, formats strfmt.Registry) error {

	if m.SubnetLicense != nil {
		if err := m.SubnetLicense.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("subnet_license")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("subnet_license")
			}
			return err
		}
	}

	return nil
}

func (m *Tenant) contextValidateTiers(ctx context.Context, formats strfmt.Registry) error {

	for i := 0; i < len(m.Tiers); i++ {

		if m.Tiers[i] != nil {
			if err := m.Tiers[i].ContextValidate(ctx, formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("tiers" + "." + strconv.Itoa(i))
				} else if ce, ok := err.(*errors.CompositeError); ok {
					return ce.ValidateName("tiers" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}

// MarshalBinary interface implementation
func (m *Tenant) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Tenant) UnmarshalBinary(b []byte) error {
	var res Tenant
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}

// TenantEndpoints tenant endpoints
//
// swagger:model TenantEndpoints
type TenantEndpoints struct {

	// console
	Console string `json:"console,omitempty"`

	// minio
	Minio string `json:"minio,omitempty"`
}

// Validate validates this tenant endpoints
func (m *TenantEndpoints) Validate(formats strfmt.Registry) error {
	return nil
}

// ContextValidate validates this tenant endpoints based on context it is used
func (m *TenantEndpoints) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *TenantEndpoints) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *TenantEndpoints) UnmarshalBinary(b []byte) error {
	var res TenantEndpoints
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
