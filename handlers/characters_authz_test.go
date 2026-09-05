package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/handlers/testutil"
)

func canViewRouter(userID int64, role string) *gin.Engine {
	return testutil.NewRouterWithUser(func(auth *gin.RouterGroup) {
		auth.GET("/characters/:id", GetCharacter)
	}, userID, role)
}

func TestCanViewCharacter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		userID   int64
		role     string
		setup    func(t *testing.T, charOwnerID int64) int64 // returns character id
		wantCode int
	}{
		{
			name:   "owner allowed",
			userID: 10, role: "user",
			setup: func(t *testing.T, ownerID int64) int64 {
				testutil.SeedCharacter(t, 101, ownerID, "Owned Char", "Human", "Fighter")
				return 101
			},
			wantCode: http.StatusOK,
		},
		{
			name:   "admin allowed for stranger's character",
			userID: 99, role: "admin",
			setup: func(t *testing.T, ownerID int64) int64 {
				testutil.SeedCharacter(t, 102, ownerID, "Other Char", "Elf", "Wizard")
				return 102
			},
			wantCode: http.StatusOK,
		},
		{
			name:   "party member allowed",
			userID: 20, role: "user",
			setup: func(t *testing.T, ownerID int64) int64 {
				// char owned by 10, in campaign 500 owned by 10, member 20
				testutil.SeedCampaign(t, 500, "Camp", "Party", 10)
				testutil.SeedCampaignMember(t, 500, 20, "player")
				testutil.SeedCharacterInCampaign(t, 103, ownerID, 500, "Party Char", "Dwarf", "Cleric")
				return 103
			},
			wantCode: http.StatusOK,
		},
		{
			name:   "stranger denied",
			userID: 30, role: "user",
			setup: func(t *testing.T, ownerID int64) int64 {
				testutil.SeedCharacter(t, 104, ownerID, "Stranger Char", "Orc", "Barbarian")
				return 104
			},
			wantCode: http.StatusForbidden,
		},
		{
			name:   "anonymous (0) denied",
			userID: 0, role: "",
			setup: func(t *testing.T, ownerID int64) int64 {
				testutil.SeedCharacter(t, 105, ownerID, "Anon Char", "Gnome", "Rogue")
				return 105
			},
			wantCode: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutil.NewDB(t)
			defer testutil.CloseDB(t)
			// seed owner for all cases (owner 10)
			testutil.SeedUser(t, 10, "owner", "user")
			if tc.userID != 0 && tc.userID != 10 {
				testutil.SeedUser(t, tc.userID, "requester", tc.role)
				if tc.role == "" {
					tc.role = "user"
				}
			}
			if tc.role == "admin" {
				testutil.SeedUser(t, tc.userID, "admin", "admin")
			}
			charID := tc.setup(t, 10)
			r := canViewRouter(tc.userID, tc.role)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/characters/"+itoaAuthz(charID), nil)
			r.ServeHTTP(w, req)
			if w.Code != tc.wantCode {
				t.Fatalf("want %d got %d body %s", tc.wantCode, w.Code, w.Body.String())
			}
		})
	}
}

func itoaAuthz(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	buf := make([]byte, 0, 20)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}
