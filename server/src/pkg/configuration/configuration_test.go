package configuration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateRevenueCatConfiguration_AllowsSelfHostedWithoutBilling(t *testing.T) {
	require.NoError(t, validateRevenueCatConfiguration(Configuration{}))
}

func TestValidateRevenueCatConfiguration_RequiresProjectIDWhenEnabled(t *testing.T) {
	err := validateRevenueCatConfiguration(Configuration{
		RevenueCatAPIKey:        "secret",
		RevenueCatEntitlementID: "entitlement",
	})
	require.ErrorContains(t, err, "revenuecat-project-id is required")
}

func TestValidateRevenueCatConfiguration_RequiresEntitlementIDWhenEnabled(t *testing.T) {
	err := validateRevenueCatConfiguration(Configuration{
		RevenueCatAPIKey:    "secret",
		RevenueCatProjectID: "project",
	})
	require.ErrorContains(t, err, "revenuecat-entitlement-id is required")
}

func TestValidateRevenueCatConfiguration_AcceptsCompleteConfig(t *testing.T) {
	require.NoError(t, validateRevenueCatConfiguration(Configuration{
		RevenueCatAPIKey:        "secret",
		RevenueCatProjectID:     "project",
		RevenueCatEntitlementID: "entitlement",
	}))
}
