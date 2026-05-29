package methods

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/nextlevelbuilder/goclaw/internal/config"
	"github.com/nextlevelbuilder/goclaw/internal/gateway"
	"github.com/nextlevelbuilder/goclaw/internal/permissions"
	"github.com/nextlevelbuilder/goclaw/internal/store"
	"github.com/nextlevelbuilder/goclaw/pkg/protocol"
)

type stubRunTimelineStore struct {
	opts  []store.RunTimelineListOpts
	items []store.RunTimelineItem
}

func (s *stubRunTimelineStore) AppendRunTimelineItem(context.Context, *store.RunTimelineItem) error {
	return nil
}

func (s *stubRunTimelineStore) ListRunTimelineItems(_ context.Context, opts store.RunTimelineListOpts) ([]store.RunTimelineItem, error) {
	s.opts = append(s.opts, opts)
	return s.items, nil
}

func TestRunTimelineGetScopesViewerByUser(t *testing.T) {
	timeline := &stubRunTimelineStore{
		items: []store.RunTimelineItem{
			{RunID: "run-1", UserID: "caller", Seq: 1, Preview: "visible"},
			{RunID: "run-1", UserID: "other", Seq: 2, Preview: "hidden"},
		},
	}
	m := NewRunTimelineMethods(timeline, &config.Config{})
	tenantID := uuid.Must(uuid.NewV7())
	client, responses := gateway.NewCapturingTestClient(permissions.RoleViewer, tenantID, "caller", 1)
	ctx := store.WithTenantID(context.Background(), tenantID)
	m.handleGet(ctx, client, sessionReqFrame(t, protocol.MethodRunTimelineGet, map[string]any{"runId": "run-1"}))

	resp := readTimelineResponse(t, responses)
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	data, ok := resp.Payload.(map[string]any)
	if !ok {
		t.Fatalf("payload type = %T", resp.Payload)
	}
	rawItems, ok := data["items"].([]any)
	if !ok {
		t.Fatalf("items type = %T", data["items"])
	}
	if len(rawItems) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(rawItems))
	}
	if timeline.opts[0].RunID != "run-1" {
		t.Fatalf("RunID = %q", timeline.opts[0].RunID)
	}
}

func readTimelineResponse(t *testing.T, ch <-chan []byte) protocol.ResponseFrame {
	t.Helper()
	raw := <-ch
	var resp protocol.ResponseFrame
	if err := json.Unmarshal(raw, &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}
