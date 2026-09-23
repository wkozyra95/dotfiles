package server

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/wkozyra95/dotfiles/api/hostd"
	"github.com/wkozyra95/dotfiles/utils/notify"
)

// Wire format of the responses, times are unix seconds.

type errorResponse struct {
	Error string `json:"error"`
	// lets the client tell a clock skew from a wrong token
	ServerTime int64 `json:"serverTime,omitempty"`
}

type statusResponse struct {
	Hostname      string   `json:"hostname"`
	UptimeSeconds int64    `json:"uptimeSeconds"`
	Features      []string `json:"features"`
}

type fileResponse struct {
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
}

type fileListResponse struct {
	Files     []fileResponse `json:"files"`
	FreeBytes uint64         `json:"freeBytes"`
}

func newFileResponse(info hostd.FileInfo) fileResponse {
	return fileResponse{Name: info.Name, Size: info.Size, Modified: info.Modified.Unix()}
}

type notifyRequest struct {
	Title   string        `json:"title"`
	Message string        `json:"message"`
	Action  *notifyAction `json:"action,omitempty"`
}

// notifyAction is the tagged union on the wire: {"type": "share-url", "url": ...}
// or {"type": "share-text", "text": ...}, decoded into the matching notify variant.
type notifyAction struct {
	notify.Action
}

func (a *notifyAction) UnmarshalJSON(data []byte) error {
	head := struct {
		Type string `json:"type"`
	}{}
	if err := json.Unmarshal(data, &head); err != nil {
		return err
	}
	switch head.Type {
	case "share-url":
		body := struct {
			Type string `json:"type"`
			URL  string `json:"url"`
		}{}
		if err := unmarshalStrict(data, &body); err != nil {
			return err
		}
		a.Action = notify.ActionOpenURL{URL: body.URL}
	case "share-text":
		body := struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{}
		if err := unmarshalStrict(data, &body); err != nil {
			return err
		}
		a.Action = notify.ActionCopyText{Text: body.Text}
	default:
		return fmt.Errorf("unsupported action type %q", head.Type)
	}
	return nil
}

// unmarshalStrict rejects fields of other variants, DisallowUnknownFields of the
// outer decoder does not reach into UnmarshalJSON.
func unmarshalStrict(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
