package nextcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// CreateUser creates a Nextcloud user through the provisioning API.
func (n *Nextcloud) CreateUser(ctx context.Context, username, password string) error {
	form := url.Values{"userid": {username}, "password": {password}}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, n.URL+"/ocs/v1.php/cloud/users", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	request.SetBasicAuth(n.AdminUsername, n.AdminPassword)
	request.Header.Set("OCS-APIRequest", "true")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("create user: HTTP %d", response.StatusCode)
	}
	return nil
}

// CreateAppPassword converts a user's login password into a new app password.
func (n *Nextcloud) CreateAppPassword(ctx context.Context, username, password string) (string, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, n.URL+"/ocs/v2.php/core/getapppassword?format=json", nil)
	if err != nil {
		return "", err
	}
	request.SetBasicAuth(username, password)
	request.Header.Set("OCS-APIRequest", "true")
	request.Header.Set("Accept", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", fmt.Errorf("create app password: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()
	var envelope struct {
		OCS struct {
			Data struct {
				AppPassword string `json:"apppassword"`
			} `json:"data"`
		} `json:"ocs"`
	}
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		return "", fmt.Errorf("decode app password: %w", err)
	}
	if response.StatusCode != http.StatusOK || envelope.OCS.Data.AppPassword == "" {
		return "", fmt.Errorf("create app password: HTTP %d", response.StatusCode)
	}
	return envelope.OCS.Data.AppPassword, nil
}
