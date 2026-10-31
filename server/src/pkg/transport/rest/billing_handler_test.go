package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"timeasy-server/pkg/subscription"
)

type enabledChecker struct{}

func (enabledChecker) IsEnabled() bool { return true }

func (enabledChecker) CheckSubscription(context.Context, string) (bool, error) { return true, nil }

func (enabledChecker) RefreshSubscription(context.Context, string) (bool, error) {
	return true, nil
}

func getBillingStatus(t *testing.T, checker subscription.Checker, webAPIKey string) map[string]any {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewBillingHandler(checker, webAPIKey)
	router.GET("/api/v1/billing/status", handler.GetBillingStatus)

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/billing/status", nil)
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
	return body
}

func TestGetBillingStatus_IncludesWebAPIKeyWhenEnabled(t *testing.T) {
	body := getBillingStatus(t, enabledChecker{}, "public-web-key")
	assert.Equal(t, true, body["enabled"])
	assert.Equal(t, "public-web-key", body["webApiKey"])
}

func TestGetBillingStatus_OmitsWebAPIKeyWithoutConfig(t *testing.T) {
	body := getBillingStatus(t, enabledChecker{}, "")
	assert.Equal(t, true, body["enabled"])
	assert.NotContains(t, body, "webApiKey")
}

func TestGetBillingStatus_OmitsWebAPIKeyWhenBillingDisabled(t *testing.T) {
	body := getBillingStatus(t, subscription.NewNoopChecker(), "public-web-key")
	assert.Equal(t, false, body["enabled"])
	assert.NotContains(t, body, "webApiKey")
}
