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

func TestValidateRateLimitConfiguration_AcceptsPositiveValues(t *testing.T) {
	require.NoError(t, validateRateLimitConfiguration(Configuration{
		RateLimitEnabled:           true,
		RateLimitRequestsPerMinute: 300,
		RateLimitBurst:             100,
	}))
}

func TestValidateRateLimitConfiguration_RejectsNonPositiveRequestsPerMinute(t *testing.T) {
	err := validateRateLimitConfiguration(Configuration{
		RateLimitEnabled:           true,
		RateLimitRequestsPerMinute: 0,
		RateLimitBurst:             100,
	})
	require.ErrorContains(t, err, "rate-limit-requests-per-minute")
}

func TestValidateRateLimitConfiguration_RejectsNonPositiveBurst(t *testing.T) {
	err := validateRateLimitConfiguration(Configuration{
		RateLimitEnabled:           true,
		RateLimitRequestsPerMinute: 300,
		RateLimitBurst:             0,
	})
	require.ErrorContains(t, err, "rate-limit-burst")
}

func TestParseCIDRList_TrimsAndDropsEmptyEntries(t *testing.T) {
	require.Equal(t, []string{"10.0.0.0/8", "192.168.0.1"}, parseCIDRList(" 10.0.0.0/8 , ,192.168.0.1 "))
	require.Empty(t, parseCIDRList(""))
}
