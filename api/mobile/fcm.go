// Package mobile sends push notifications to the phones registered with hostd
// (POST /push-token) through Firebase Cloud Messaging.
package mobile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2/jwt"

	"github.com/wkozyra95/dotfiles/logger"
)

var log = logger.NamedLogger("mobile")

const (
	fcmScope   = "https://www.googleapis.com/auth/firebase.messaging"
	fcmSendURL = "https://fcm.googleapis.com/v1/projects/%s/messages:send"
	// The data payload is limited to 4 KB and the text is in it twice (message
	// and text), the rest is for the other keys and the JSON framing.
	fcmMaxTextLen = 1900
)

var (
	// ErrUnregistered means the token is no longer valid (app reinstalled, token
	// rotated), the device has to register again.
	ErrUnregistered = errors.New("device is not registered with FCM anymore")
	ErrInvalidText  = errors.New("invalid text")
)

// ServiceAccount is the key of a Firebase service account, the JSON downloaded
// from the Firebase console (Project settings -> Service accounts).
type ServiceAccount struct {
	ProjectID    string `json:"project_id"`
	ClientEmail  string `json:"client_email"`
	PrivateKey   string `json:"private_key"`
	PrivateKeyID string `json:"private_key_id"`
	TokenURI     string `json:"token_uri"`
}

func ParseServiceAccount(raw []byte) (ServiceAccount, error) {
	account := ServiceAccount{}
	if err := json.Unmarshal(raw, &account); err != nil {
		return ServiceAccount{}, fmt.Errorf("service account: %w", err)
	}
	if account.ProjectID == "" || account.ClientEmail == "" || account.PrivateKey == "" {
		return ServiceAccount{}, errors.New("service account: project_id, client_email and private_key are required")
	}
	if account.TokenURI == "" {
		account.TokenURI = "https://oauth2.googleapis.com/token"
	}
	return account, nil
}

// Client sends messages through the FCM HTTP v1 API, authenticated with an
// OAuth2 token obtained for the service account.
type Client struct {
	projectID string
	http      *http.Client
	// fcmSendURL, replaced in tests
	sendURL string
}

func NewClient(ctx context.Context, account ServiceAccount) *Client {
	config := &jwt.Config{
		Email:        account.ClientEmail,
		PrivateKey:   []byte(account.PrivateKey),
		PrivateKeyID: account.PrivateKeyID,
		Scopes:       []string{fcmScope},
		TokenURL:     account.TokenURI,
	}
	client := config.Client(ctx)
	client.Timeout = 30 * time.Second
	return &Client{projectID: account.ProjectID, http: client, sendURL: fcmSendURL}
}

// fcmMessage is data-only on purpose: with a notification block the Firebase
// SDK shows the message itself while the app is in the background and the app
// never sees it if it is swiped away. Data-only messages always reach the app
// (expo-notifications presents them from the data keys) and it stores them.
type fcmMessage struct {
	Token   string            `json:"token"`
	Data    map[string]string `json:"data"`
	Android fcmAndroid        `json:"android"`
}

type fcmAndroid struct {
	// high is needed for a data-only message to wake the app in doze
	Priority string `json:"priority"`
}

type fcmErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
		Details []struct {
			ErrorCode string `json:"errorCode"`
		} `json:"details"`
	} `json:"error"`
}

// SendText delivers the text to the device with that token. The keys of the
// data are what the myremote app expects: title and message are shown as the
// notification, text is what the app copies on tap, channelId is the
// notification channel the app creates.
func (c *Client) SendText(ctx context.Context, token string, title string, text string) error {
	if text == "" || len(text) > fcmMaxTextLen {
		return fmt.Errorf("%w: has to be 1-%d bytes", ErrInvalidText, fcmMaxTextLen)
	}
	message := fcmMessage{
		Token: token,
		Data: map[string]string{
			"type":      "text",
			"title":     title,
			"message":   text,
			"text":      text,
			"channelId": "default",
		},
		Android: fcmAndroid{Priority: "high"},
	}
	body, err := json.Marshal(struct {
		Message fcmMessage `json:"message"`
	}{Message: message})
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf(c.sendURL, c.projectID),
		bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 64<<10))
	if response.StatusCode == http.StatusOK {
		log.Debugf("FCM response: %s", responseBody)
		return nil
	}
	return fcmError(response.StatusCode, responseBody)
}

func fcmError(status int, body []byte) error {
	parsed := fcmErrorResponse{}
	if err := json.Unmarshal(body, &parsed); err != nil || parsed.Error.Message == "" {
		return fmt.Errorf("FCM responded with %d: %s", status, bytes.TrimSpace(body))
	}
	for _, detail := range parsed.Error.Details {
		if detail.ErrorCode == "UNREGISTERED" {
			return ErrUnregistered
		}
	}
	if status == http.StatusNotFound {
		return ErrUnregistered
	}
	return fmt.Errorf("FCM responded with %d %s: %s", status, parsed.Error.Status, parsed.Error.Message)
}
