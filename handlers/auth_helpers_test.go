package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func TestMustGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", int64(42))
	uid, ok := MustGetUserID(c)
	if !ok || uid != 42 {
		t.Fatalf("expected 42 got %d ok=%v", uid, ok)
	}
	c2, _ := gin.CreateTestContext(w)
	c2.Request = httptest.NewRequest("GET", "/", nil)
	if _, ok := MustGetUserID(c2); ok {
		t.Fatal("expected false when missing")
	}
}

func TestParseCampaignID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "123"}}
	id, ok := ParseCampaignID(c)
	if !ok || id != 123 {
		t.Fatalf("expected 123 got %d", id)
	}
	c.Params = gin.Params{{Key: "id", Value: "bad"}}
	if _, ok := ParseCampaignID(c); ok {
		t.Fatal("expected false on bad")
	}
}

func TestIsCampaignMember(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "owner", "user")
	testutil.SeedUser(t, 2, "member", "user")
	testutil.SeedUser(t, 3, "outsider", "user")
	testutil.SeedCampaign(t, 1, "Camp", "Party", 1)
	testutil.SeedCampaignMember(t, 1, 2, "player")

	ctx := t.Context()
	if ok, _ := IsCampaignMember(ctx, 1, 1); !ok {
		t.Fatal("owner should be member")
	}
	if ok, _ := IsCampaignMember(ctx, 1, 2); !ok {
		t.Fatal("member should be member")
	}
	if ok, _ := IsCampaignMember(ctx, 1, 3); ok {
		t.Fatal("outsider should not be member")
	}
}
