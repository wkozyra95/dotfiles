package mobile

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseServiceAccount(t *testing.T) {
	account, err := ParseServiceAccount([]byte(`{
		"type": "service_account",
		"project_id": "my-project",
		"private_key_id": "abc",
		"private_key": "-----BEGIN PRIVATE KEY-----\nMII\n-----END PRIVATE KEY-----\n",
		"client_email": "firebase-adminsdk@my-project.iam.gserviceaccount.com",
		"client_id": "123"
	}`))
	assert.Nil(t, err)
	assert.Equal(t, "my-project", account.ProjectID)
	assert.Equal(t, "https://oauth2.googleapis.com/token", account.TokenURI)

	_, err = ParseServiceAccount([]byte(`{"project_id": "my-project"}`))
	assert.NotNil(t, err)
	_, err = ParseServiceAccount([]byte(`null`))
	assert.NotNil(t, err)
	_, err = ParseServiceAccount([]byte(`not json`))
	assert.NotNil(t, err)
}

func TestFcmError(t *testing.T) {
	unregistered := `{"error":{"code":404,"message":"Requested entity was not found.","status":"NOT_FOUND",
		"details":[{"@type":"type.googleapis.com/google.firebase.fcm.v1.FcmError","errorCode":"UNREGISTERED"}]}}`
	assert.ErrorIs(t, fcmError(http.StatusNotFound, []byte(unregistered)), ErrUnregistered)

	invalid := `{"error":{"code":400,"message":"The registration token is not a valid FCM registration token",
		"status":"INVALID_ARGUMENT","details":[{"@type":"type.googleapis.com/google.firebase.fcm.v1.FcmError","errorCode":"INVALID_ARGUMENT"}]}}`
	err := fcmError(http.StatusBadRequest, []byte(invalid))
	assert.NotErrorIs(t, err, ErrUnregistered)
	assert.Contains(t, err.Error(), "INVALID_ARGUMENT")

	err = fcmError(http.StatusBadGateway, []byte("<html>bad gateway</html>"))
	assert.Equal(t, "FCM responded with 502: <html>bad gateway</html>", err.Error())
}

func testServiceAccountKey(t *testing.T) string {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	assert.Nil(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(key)
	assert.Nil(t, err)
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

// TestSendText runs the token exchange and the send against a fake FCM.
func TestSendText(t *testing.T) {
	sent := []map[string]any{}
	authorizations := []string{}
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			assert.Nil(t, r.ParseForm())
			assert.Equal(t, "urn:ietf:params:oauth:grant-type:jwt-bearer", r.Form.Get("grant_type"))
			assert.NotEmpty(t, r.Form.Get("assertion"))
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"fake-access-token","token_type":"Bearer","expires_in":3600}`))
		case "/v1/projects/my-project/messages:send":
			authorizations = append(authorizations, r.Header.Get("Authorization"))
			body := map[string]any{}
			assert.Nil(t, json.NewDecoder(r.Body).Decode(&body))
			message := body["message"].(map[string]any)
			sent = append(sent, message)
			if message["token"] == "gone" {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write(
					[]byte(`{"error":{"code":404,"message":"Requested entity was not found.","status":"NOT_FOUND",
					"details":[{"@type":"type.googleapis.com/google.firebase.fcm.v1.FcmError","errorCode":"UNREGISTERED"}]}}`),
				)
				return
			}
			_, _ = w.Write([]byte(`{"name":"projects/my-project/messages/1"}`))
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer fake.Close()

	account := ServiceAccount{
		ProjectID:   "my-project",
		ClientEmail: "sa@my-project.iam.gserviceaccount.com",
		PrivateKey:  testServiceAccountKey(t),
		TokenURI:    fake.URL + "/token",
	}
	client := NewClient(context.Background(), account)
	client.sendURL = fake.URL + "/v1/projects/%s/messages:send"

	assert.Nil(t, client.SendText(context.Background(), "phone-token", "home", "hello ą"))
	assert.ErrorIs(t, client.SendText(context.Background(), "gone", "home", "hello"), ErrUnregistered)
	assert.ErrorIs(t, client.SendText(context.Background(), "phone-token", "home", ""), ErrInvalidText)

	assert.Equal(t, []string{"Bearer fake-access-token", "Bearer fake-access-token"}, authorizations)
	assert.Equal(t, 2, len(sent))
	assert.Equal(t, "phone-token", sent[0]["token"])
	// data-only, a notification block would be presented by the SDK instead of the app
	_, hasNotification := sent[0]["notification"]
	assert.False(t, hasNotification)
	assert.Equal(t, map[string]any{
		"type":      "text",
		"title":     "home",
		"message":   "hello ą",
		"text":      "hello ą",
		"channelId": "default",
	}, sent[0]["data"])
	assert.Equal(t, map[string]any{"priority": "high"}, sent[0]["android"])
}

func TestSendTextLength(t *testing.T) {
	client := NewClient(context.Background(), ServiceAccount{PrivateKey: "unused"})
	// rejected before anything is sent, the client has no reachable endpoint
	assert.ErrorIs(
		t,
		client.SendText(context.Background(), "phone-token", "home", strings.Repeat("a", fcmMaxTextLen+1)),
		ErrInvalidText,
	)
	// the whole data payload stays under the 4 KB limit of FCM at the longest text
	data := map[string]string{
		"type": "text", "title": strings.Repeat("h", 64), "message": strings.Repeat("a", fcmMaxTextLen),
		"text": strings.Repeat("a", fcmMaxTextLen), "channelId": "default",
	}
	encoded, err := json.Marshal(data)
	assert.Nil(t, err)
	assert.Less(t, len(encoded), 4096)
}
