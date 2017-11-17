// Copyright 2017 The Cockroach Authors.
//
// Licensed as a CockroachDB Enterprise file under the Cockroach Community
// License (the "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
//     https://github.com/cockroachdb/cockroach/blob/master/licenses/CCL.txt

package utilccl

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/utilccl/licenseccl"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

var enterpriseLicense = settings.RegisterValidatedStringSetting(
	"enterprise.license",
	"the encoded cluster license",
	"",
	func(s string) error {
		_, err := licenseccl.Decode(s)
		return err
	},
)

func init() {
	enterpriseLicense.Hide()
}

var testingEnterpriseEnabled = false

// TestingEnableEnterprise allows overriding the license check in tests.
func TestingEnableEnterprise() func() {
	before := testingEnterpriseEnabled
	testingEnterpriseEnabled = true
	return func() {
		testingEnterpriseEnabled = before
	}
}

// CheckEnterpriseEnabled returns a non-nil error if the requested enterprise
// feature is not enabled, including information or a link explaining how to
// enable it.
func CheckEnterpriseEnabled(st *cluster.Settings, cluster uuid.UUID, org, feature string) error {
	if testingEnterpriseEnabled {
		return nil
	}
	return checkEnterpriseEnabledAt(st, timeutil.Now(), cluster, org, feature)
}

func init() {
	server.LicenseCheckFn = CheckEnterpriseEnabled
}

func checkEnterpriseEnabledAt(
	st *cluster.Settings, at time.Time, cluster uuid.UUID, org, feature string,
) error {
	var lic *licenseccl.License
	// FIXME(tschottdorf): see whether it makes sense to cache the decoded
	// license.
	if str := enterpriseLicense.Get(&st.SV); str != "" {
		var err error
		if lic, err = licenseccl.Decode(str); err != nil {
			return err
		}
	}
	return lic.Check(at, cluster, org, feature)
}
