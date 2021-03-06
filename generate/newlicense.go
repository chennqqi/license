/*
Copyright 2017 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package generate

import (
	"time"

	"github.com/gravitational/license"
	"github.com/gravitational/license/authority"
	"github.com/gravitational/license/constants"

	"github.com/cloudflare/cfssl/csr"
	"github.com/gravitational/trace"
)

// NewLicenseInfo encapsulates fields needed to generate a license
type NewLicenseInfo struct {
	// MaxNodes is maximum number of nodes the license allows
	MaxNodes int
	// ValidFor is validity period for the license
	ValidFor time.Duration
	// StopApp indicates whether the app should be stopped when the license expires
	StopApp bool
	// CustomerName is the name of the customer the license is generated for
	CustomerName string
	// CustomerName is the email of the customer the license is generated for
	CustomerEmail string
	// CustomerMetadata is arbitrary metadata to add to the license
	CustomerMetadata string
	// ProductName is the name of the product the license is for
	ProductName string
	// ProductVersion is product version the license is for
	ProductVersion string
	// AccountID is the id of the account the license is for
	AccountID string
	// EncryptionKey is the passphrase for decoding encrypted packages
	EncryptionKey []byte
	// TLSKeyPair is the certificate authority to sign the license with
	TLSKeyPair authority.TLSKeyPair
}

// Check checks the new license request
func (i *NewLicenseInfo) Check() error {
	if i.MaxNodes < 1 {
		return trace.BadParameter("maximum number of servers must be 1 or more")
	}
	if time.Now().Add(i.ValidFor).Before(time.Now()) {
		return trace.BadParameter("expiration date can't be in the past")
	}
	if len(i.TLSKeyPair.CertPEM) == 0 {
		return trace.BadParameter("certificate authority must be provided")
	}
	return nil
}

// NewLicense generates a new license according to the provided request
func NewLicense(info NewLicenseInfo) (string, error) {
	if err := info.Check(); err != nil {
		return "", trace.Wrap(err)
	}
	certificateBytes, err := newCertificate(info)
	if err != nil {
		return "", trace.Wrap(err)
	}
	return string(certificateBytes), nil
}

// NewTestLicense generates a new license for use in tests
func NewTestLicense() (*license.License, error) {
	ca, err := authority.GenerateSelfSignedCA(csr.CertificateRequest{
		CN: constants.LicenseKeyPair,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	lic, err := NewLicense(NewLicenseInfo{
		MaxNodes:   3,
		ValidFor:   time.Duration(time.Hour),
		TLSKeyPair: *ca,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	parsed, err := license.ParseString(lic)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return parsed, nil
}
