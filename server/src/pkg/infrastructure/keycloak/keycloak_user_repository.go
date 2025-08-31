package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"timeasy-server/pkg/domain/model"
	"timeasy-server/pkg/domain/repository"

	"github.com/gofrs/uuid"
)

type keycloakUserRepository struct {
	keycloakURL string
	realm       string
	httpClient  *http.Client
}

type KeycloakUser struct {
	ID         string              `json:"id"`
	Username   string              `json:"username"`
	Email      string              `json:"email"`
	FirstName  string              `json:"firstName"`
	LastName   string              `json:"lastName"`
	Enabled    bool                `json:"enabled"`
	Attributes map[string][]string `json:"attributes,omitempty"`
}

type KeycloakPasswordUpdate struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
	Confirmation    string `json:"confirmation"`
}

func NewKeycloakUserRepository(keycloakURL, realm string) repository.UserRepository {
	return &keycloakUserRepository{
		keycloakURL: keycloakURL,
		realm:       realm,
		httpClient:  &http.Client{},
	}
}

func (r *keycloakUserRepository) GetUserProfile(ctx context.Context, userToken string) (*model.User, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/userinfo", r.keycloakURL, r.realm)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Accept", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("keycloak returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return r.mapUserInfoToModel(userInfo), nil
}

func (r *keycloakUserRepository) mapUserInfoToModel(userInfo map[string]interface{}) *model.User {
	user := &model.User{
		Username:    getString(userInfo, "preferred_username", ""),
		Email:       getString(userInfo, "email", ""),
		FirstName:   getString(userInfo, "given_name", ""),
		LastName:    getString(userInfo, "family_name", ""),
		DisplayName: getString(userInfo, "name", ""),
		Attributes:  make(map[string]string),
	}

	if idStr, exists := userInfo["sub"]; exists {
		if id, err := uuid.FromString(idStr.(string)); err == nil {
			user.ID = id
		}
	}

	if user.DisplayName == "" {
		user.DisplayName = strings.TrimSpace(user.FirstName + " " + user.LastName)
		if user.DisplayName == "" {
			user.DisplayName = user.Username
		}
	}

	if locale, exists := userInfo["locale"]; exists {
		if localeStr, ok := locale.(string); ok {
			user.Language = &localeStr
		}
	}

	return user
}

func (r *keycloakUserRepository) UpdateUserProfile(ctx context.Context, userToken string, updateRequest *model.UserProfileUpdateRequest) (*model.User, error) {
	url := fmt.Sprintf("%s/realms/%s/account", r.keycloakURL, r.realm)

	currentUser, err := r.GetUserProfile(ctx, userToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// Prepare the update payload using the correct format for Keycloak Account API
	updatePayload := map[string]interface{}{
		"username":  currentUser.Username,
		"firstName": currentUser.FirstName,
		"lastName":  currentUser.LastName,
		"email":     currentUser.Email,
	}

	// Apply updates - only include fields that are being changed
	if updateRequest.FirstName != nil {
		updatePayload["firstName"] = *updateRequest.FirstName
	}
	if updateRequest.LastName != nil {
		updatePayload["lastName"] = *updateRequest.LastName
	}
	if updateRequest.Email != nil {
		updatePayload["email"] = *updateRequest.Email
	}

	attributes := make(map[string][]string)

	// Copy existing attributes from current user
	for key, value := range currentUser.Attributes {
		attributes[key] = []string{value}
	}

	// Add language preference to attributes if provided
	if updateRequest.Language != nil {
		attributes["locale"] = []string{*updateRequest.Language}
	}

	// Add other custom attributes if provided
	for key, value := range updateRequest.Attributes {
		attributes[key] = []string{value}
	}

	// Only add attributes if we have any
	if len(attributes) > 0 {
		updatePayload["attributes"] = attributes
	}

	jsonData, err := json.Marshal(updatePayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal update request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		bodyBytes = []byte("could not read response body")
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("keycloak account API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return r.GetUserProfile(ctx, userToken)
}

func (r *keycloakUserRepository) ChangePassword(ctx context.Context, userToken string, passwordRequest *model.PasswordChangeRequest) error {
	url := fmt.Sprintf("%s/realms/%s/account/credentials/password", r.keycloakURL, r.realm)

	passwordUpdate := KeycloakPasswordUpdate{
		CurrentPassword: passwordRequest.CurrentPassword,
		NewPassword:     passwordRequest.NewPassword,
		Confirmation:    passwordRequest.NewPassword,
	}

	jsonData, err := json.Marshal(passwordUpdate)
	if err != nil {
		return fmt.Errorf("failed to marshal password update: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+userToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("keycloak password change failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (r *keycloakUserRepository) GetOrCreateUser(ctx context.Context, keycloakUserID uuid.UUID, keycloakData map[string]interface{}) (*model.User, error) {
	user := &model.User{
		ID:          keycloakUserID,
		Username:    getString(keycloakData, "preferred_username", ""),
		Email:       getString(keycloakData, "email", ""),
		FirstName:   getString(keycloakData, "given_name", ""),
		LastName:    getString(keycloakData, "family_name", ""),
		DisplayName: getString(keycloakData, "name", ""),
		Attributes:  make(map[string]string),
	}

	if user.DisplayName == "" {
		user.DisplayName = strings.TrimSpace(user.FirstName + " " + user.LastName)
		if user.DisplayName == "" {
			user.DisplayName = user.Username
		}
	}

	return user, nil
}

func getString(data map[string]interface{}, key, defaultValue string) string {
	if value, exists := data[key]; exists {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return defaultValue
}

func (r *keycloakUserRepository) mapKeycloakUserToModel(keycloakUser *KeycloakUser) *model.User {
	user := &model.User{
		Username:    keycloakUser.Username,
		Email:       keycloakUser.Email,
		FirstName:   keycloakUser.FirstName,
		LastName:    keycloakUser.LastName,
		DisplayName: fmt.Sprintf("%s %s", keycloakUser.FirstName, keycloakUser.LastName),
		Attributes:  make(map[string]string),
	}

	if id, err := uuid.FromString(keycloakUser.ID); err == nil {
		user.ID = id
	}

	if lang, exists := keycloakUser.Attributes["language"]; exists && len(lang) > 0 {
		user.Language = &lang[0]
	}

	for key, values := range keycloakUser.Attributes {
		if len(values) > 0 {
			user.Attributes[key] = values[0]
		}
	}

	user.DisplayName = strings.TrimSpace(user.DisplayName)
	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}

	return user
}
