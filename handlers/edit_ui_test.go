package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/handlers/testutil"
)

func putForm(t *testing.T, r http.Handler, path string, data map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	form := url.Values{}
	for k, v := range data {
		form.Set(k, v)
	}
	req := httptest.NewRequest(http.MethodPut, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestUpdatePartyItem(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCampaign(t, 1, "Camp", "The Party", 1)
	res, err := db.DB.Exec("INSERT INTO party_items(campaign_id, name, quantity, notes) VALUES(1,'Rope',1,'')")
	if err != nil {
		t.Fatalf("seed item: %v", err)
	}
	id, _ := res.LastInsertId()

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.PUT("/party-items/:id", UpdateCampaignPartyItem)
	})
	w := testutil.PutJSON(t, r, fmt.Sprintf("/api/party-items/%d", id), map[string]any{
		"name": "Silken Rope", "quantity": 5, "notes": "50 ft",
	})
	testutil.AssertStatus(t, w, http.StatusOK)

	var name, notes string
	var qty int
	db.DB.QueryRow("SELECT name, quantity, notes FROM party_items WHERE id=?", id).Scan(&name, &qty, &notes)
	if name != "Silken Rope" || qty != 5 || notes != "50 ft" {
		t.Fatalf("update not persisted: name=%q qty=%d notes=%q", name, qty, notes)
	}
}

func TestHtmxFeatureAndProficiencyEdit(t *testing.T) {
	testutil.NewDB(t)
	defer testutil.CloseDB(t)
	testutil.SeedUser(t, 1, "admin", "admin")
	testutil.SeedCharacter(t, 10, 1, "Hero", "Human", "Fighter")
	fres, err := db.DB.Exec("INSERT INTO character_features(character_id,name,description,source,level_gained) VALUES(10,'Old Feature','d','src',1)")
	if err != nil {
		t.Fatalf("seed feature: %v", err)
	}
	fid, _ := fres.LastInsertId()
	pres, err := db.DB.Exec("INSERT INTO character_proficiencies(character_id,type,name) VALUES(10,'skill','Stealth')")
	if err != nil {
		t.Fatalf("seed proficiency: %v", err)
	}
	pid, _ := pres.LastInsertId()

	r := testutil.NewRouter(func(auth *gin.RouterGroup) {
		auth.GET("/htmx/features/:id/edit", HtmxEditFeatureForm)
		auth.PUT("/htmx/features/:id", HtmxUpdateFeature)
		auth.GET("/htmx/proficiencies/:id/edit", HtmxEditProficiencyForm)
		auth.PUT("/htmx/proficiencies/:id", HtmxUpdateProficiency)
	})

	// Edit form prefills the existing values.
	w := testutil.Get(t, r, fmt.Sprintf("/api/htmx/features/%d/edit", fid))
	testutil.AssertStatus(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `value="Old Feature"`) {
		t.Fatalf("feature edit form missing prefilled name: %s", w.Body.String())
	}

	// Update feature.
	w = putForm(t, r, fmt.Sprintf("/api/htmx/features/%d", fid), map[string]string{
		"character_id": "10", "name": "New Feature", "description": "desc", "source": "PHB", "level_gained": "3",
	})
	testutil.AssertStatus(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), "New Feature") {
		t.Fatalf("feature update not reflected: %s", w.Body.String())
	}
	var fname, fsource string
	var flevel int
	db.DB.QueryRow("SELECT name, source, level_gained FROM character_features WHERE id=?", fid).Scan(&fname, &fsource, &flevel)
	if fname != "New Feature" || fsource != "PHB" || flevel != 3 {
		t.Fatalf("feature not updated: %q %q %d", fname, fsource, flevel)
	}

	// Edit form + update for proficiency.
	w = testutil.Get(t, r, fmt.Sprintf("/api/htmx/proficiencies/%d/edit", pid))
	testutil.AssertStatus(t, w, http.StatusOK)
	if !strings.Contains(w.Body.String(), `value="Stealth"`) {
		t.Fatalf("proficiency edit form missing prefilled name: %s", w.Body.String())
	}
	w = putForm(t, r, fmt.Sprintf("/api/htmx/proficiencies/%d", pid), map[string]string{
		"character_id": "10", "type": "tool", "name": "Thieves' Tools",
	})
	testutil.AssertStatus(t, w, http.StatusOK)
	var ptype, pname string
	db.DB.QueryRow("SELECT type, name FROM character_proficiencies WHERE id=?", pid).Scan(&ptype, &pname)
	if ptype != "tool" || pname != "Thieves' Tools" {
		t.Fatalf("proficiency not updated: %q %q", ptype, pname)
	}
}
