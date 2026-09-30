package nextcloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// SharePermissions defines public-link row permissions.
type SharePermissions struct {
	Read   bool
	Create bool
	Update bool
	Delete bool
}

// CreateShare creates a public table share and adjusts its row permissions.
func (n *Nextcloud) CreateShare(ctx context.Context, username, password string, tableID int64, sharePassword string, permissions SharePermissions) (string, error) {
	body := map[string]any{}
	if sharePassword != "" {
		body["password"] = sharePassword
	}
	data, _ := json.Marshal(body)
	path := fmt.Sprintf("/ocs/v2.php/apps/tables/api/2/tables/%d/share", tableID)
	var created struct {
		ShareToken string `json:"shareToken"`
	}
	if err := n.doOCS(ctx, username, password, http.MethodPost, path, data, &created); err != nil {
		return "", err
	}
	if permissions == (SharePermissions{Read: true}) {
		return created.ShareToken, nil
	}

	var shares []struct {
		ID           int64  `json:"id"`
		Token        string `json:"token"`
		ReceiverType string `json:"receiverType"`
	}
	if err := n.doJSON(ctx, username, password, http.MethodGet, fmt.Sprintf("/index.php/apps/tables/share/table/%d", tableID), nil, &shares); err != nil {
		return "", err
	}
	var shareID int64
	for _, share := range shares {
		if share.Token == created.ShareToken || (share.ReceiverType == "link" && share.Token != "") {
			shareID = share.ID
			break
		}
	}
	if shareID == 0 {
		return "", fmt.Errorf("created link share was not listed")
	}
	update, _ := json.Marshal(map[string]bool{
		"permissionRead": permissions.Read, "permissionCreate": permissions.Create,
		"permissionUpdate": permissions.Update, "permissionDelete": permissions.Delete,
	})
	if err := n.doJSON(ctx, username, password, http.MethodPut, fmt.Sprintf("/index.php/apps/tables/share/%d/permissions", shareID), update, nil); err != nil {
		return "", err
	}
	return created.ShareToken, nil
}

func (n *Nextcloud) doOCS(ctx context.Context, username, password, method, path string, body []byte, target any) error {
	var envelope struct {
		OCS struct {
			Meta struct {
				StatusCode int    `json:"statuscode"`
				Message    string `json:"message"`
			} `json:"meta"`
			Data json.RawMessage `json:"data"`
		} `json:"ocs"`
	}
	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}
	if err := n.doJSON(ctx, username, password, method, path+separator+"format=json", body, &envelope); err != nil {
		return err
	}
	if envelope.OCS.Meta.StatusCode != 100 && envelope.OCS.Meta.StatusCode != 200 {
		return fmt.Errorf("OCS error %d: %s", envelope.OCS.Meta.StatusCode, envelope.OCS.Meta.Message)
	}
	if target == nil || len(envelope.OCS.Data) == 0 {
		return nil
	}
	return json.Unmarshal(envelope.OCS.Data, target)
}

func (n *Nextcloud) doJSON(ctx context.Context, username, password, method, path string, body []byte, target any) error {
	request, err := http.NewRequestWithContext(ctx, method, n.URL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.SetBasicAuth(username, password)
	request.Header.Set("OCS-APIRequest", "true")
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("request %s: HTTP %d", path, response.StatusCode)
	}
	if target == nil {
		return nil
	}
	return json.NewDecoder(response.Body).Decode(target)
}
