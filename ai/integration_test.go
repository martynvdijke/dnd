package ai

import (
	"context"
	"os"
	"testing"

	"villum/db"
)

func TestAIIntegration(t *testing.T) {
	path := "/tmp/villum_ai_integ.db"
	os.Remove(path)
	if err := db.Init(path); err != nil {
		t.Fatalf("db init: %v", err)
	}
	db.Seed()
	defer func() { db.Close(); os.Remove(path) }()
	ctx := context.Background()
	_ = IsEnabled(ctx, db.DB)
	// set disabled
	db.DB.Exec("INSERT OR REPLACE INTO app_settings(key,value) VALUES('ai_enabled','0')")
	_ = IsEnabled(ctx, db.DB)
	db.DB.Exec("INSERT OR REPLACE INTO app_settings(key,value) VALUES('ai_enabled','1')")
	_ = SystemPromptForCompendiumQA("block")
	_ = SystemPromptForCampaignQA("block")
	_, _ = ResolveEndpoint(ctx, db.DB, 0)
	_ = SanitizeError(os.ErrInvalid)
	_ = TruncateResponse("short")
	_ = TruncateResponse(string(make([]byte, 500)))
	_ = NormalizeAIBaseURL(" https://example.com/v1/ ")
	_, _ = ValidateAIBaseURL("https://example.com/v1")
	_, _, _ = GenerateText(ctx, db.DB, GenerateRequest{Prompt: ""})
	_, _, _ = GenerateText(ctx, db.DB, GenerateRequest{Prompt: "hi", EndpointID: 9999})
	_, _, _ = GenerateChat(ctx, db.DB, 0, nil, nil, "", TestTimeout)
	_, _, _ = GenerateChat(ctx, db.DB, 1, []map[string]string{{"role": "user", "content": "hi"}}, nil, "", TestTimeout)
}
