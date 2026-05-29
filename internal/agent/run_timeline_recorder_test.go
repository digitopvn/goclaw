package agent

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/nextlevelbuilder/goclaw/internal/store"
	"github.com/nextlevelbuilder/goclaw/pkg/protocol"
)

func TestRunTimelineItemFromEventScrubsToolArguments(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV7())
	item, ok := runTimelineItemFromEvent(AgentEvent{
		Type:       protocol.AgentEventToolCall,
		AgentID:    "default",
		RunID:      "run-1",
		UserID:     "user-1",
		Channel:    "web",
		ChatID:     "chat-1",
		SessionKey: "session-1",
		TenantID:   tenantID,
		Payload: map[string]any{
			"name":      "exec_command",
			"id":        "call-1",
			"arguments": map[string]any{"cmd": "echo sk-abcdefghijklmnopqrstuvwxyz123456"},
		},
	}, 7)
	if !ok {
		t.Fatal("expected timeline item")
	}
	if item.ItemType != store.RunTimelineItemTypeToolCall {
		t.Fatalf("ItemType = %q", item.ItemType)
	}
	if item.Seq != 7 {
		t.Fatalf("Seq = %d, want 7", item.Seq)
	}
	if item.ToolName != "exec_command" || item.ToolCallID != "call-1" {
		t.Fatalf("tool fields = %q/%q", item.ToolName, item.ToolCallID)
	}
	if strings.Contains(item.Preview, "sk-abcdefghijklmnopqrstuvwxyz123456") {
		t.Fatalf("preview leaked secret: %s", item.Preview)
	}
	if !strings.Contains(item.Preview, "[REDACTED]") {
		t.Fatalf("preview missing redaction: %s", item.Preview)
	}
	if item.AgentID != nil {
		t.Fatalf("AgentID = %v, want nil for agent key", item.AgentID)
	}
	if !strings.Contains(string(item.Metadata), `"agent_key":"default"`) {
		t.Fatalf("metadata missing agent_key: %s", item.Metadata)
	}
}

func TestRunTimelineItemFromEventDropsUnsupportedAndThinking(t *testing.T) {
	tenantID := uuid.Must(uuid.NewV7())
	if _, ok := runTimelineItemFromEvent(AgentEvent{
		Type:       protocol.ChatEventThinking,
		RunID:      "run-1",
		SessionKey: "session-1",
		TenantID:   tenantID,
	}, 1); ok {
		t.Fatal("thinking event should not be archived")
	}

	item, ok := runTimelineItemFromEvent(AgentEvent{
		Type:       protocol.AgentEventRunCompleted,
		RunID:      "run-1",
		SessionKey: "session-1",
		TenantID:   tenantID,
		Payload: map[string]any{
			"content":  "visible <thinking>hidden chain</thinking> done",
			"thinking": "raw hidden chain",
		},
	}, 1)
	if !ok {
		t.Fatal("expected completed item")
	}
	if strings.Contains(item.Preview, "hidden chain") || strings.Contains(item.Preview, "raw hidden") {
		t.Fatalf("preview leaked thinking: %q", item.Preview)
	}
	if item.Preview != "visible  done" {
		t.Fatalf("Preview = %q", item.Preview)
	}
}
