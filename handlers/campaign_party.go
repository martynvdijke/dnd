package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"villum/db"
	"villum/ent"
	"villum/ent/campaign"
	"villum/ent/campaignmember"
	"villum/ent/character"
	"villum/ent/characterspellcasting"
)

type PartyMember struct {
	ID            int64  `json:"id"`
	UserID        int64  `json:"user_id"`
	OwnerName     string `json:"owner_name"`
	Name          string `json:"name"`
	Race          string `json:"race"`
	RaceColor     string `json:"race_color"`
	Class         string `json:"class"`
	Level         int    `json:"level"`
	AC            int    `json:"ac"`
	HPMax         int    `json:"hp_max"`
	HPCurrent     int    `json:"hp_current"`
	TempHP        int    `json:"temp_hp"`
	Status        string `json:"status"`
	PortraitURL   string `json:"portrait_url"`
	CampaignID    *int64 `json:"campaign_id"`
	CharacterType string `json:"character_type"`
	DMNotes       string `json:"dm_notes,omitempty"`
	Owned         bool   `json:"owned"`
}

type CampaignGroup struct {
	ID        int64         `json:"id"`
	Name      string        `json:"name"`
	PartyName string        `json:"party_name"`
	OwnerName string        `json:"owner_name"`
	Members   []PartyMember `json:"members"`
}

func GetPartyView(c *gin.Context) {
	userID, _ := c.Get("user_id")
	currentUID := userID.(int64)
	role, _ := c.Get("role")
	ctx := c.Request.Context()
	var rcMap map[string]string
	var camps []*ent.Campaign
	var err error
	if role == "admin" {
		camps, err = db.Client.Campaign.Query().WithUser().All(ctx)
	} else {
		camps, err = db.Client.Campaign.Query().Where(campaign.Or(campaign.UserID(currentUID), campaign.HasMembersWith(campaignmember.UserID(currentUID)))).WithUser().All(ctx)
	}
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	includeAll := role == "admin"
	userSet := make(map[int64]bool)
	userSet[currentUID] = true
	for _, ca := range camps {
		userSet[ca.UserID] = true
		if !includeAll {
			members, _ := db.Client.CampaignMember.Query().Where(campaignmember.CampaignID(ca.ID)).All(ctx)
			for _, m := range members {
				userSet[m.UserID] = true
			}
		}
	}
	uidList := make([]int64, 0, len(userSet))
	for uid := range userSet {
		uidList = append(uidList, uid)
	}
	dmCampaignIDs := make(map[int64]bool)
	if role != "admin" {
		for _, ca := range camps {
			if ca.UserID == currentUID {
				dmCampaignIDs[ca.ID] = true
				continue
			}
			n, err := db.Client.CampaignMember.Query().Where(campaignmember.CampaignIDEQ(ca.ID), campaignmember.UserIDEQ(currentUID), campaignmember.RoleEQ("dm")).Count(ctx)
			if err == nil && n > 0 {
				dmCampaignIDs[ca.ID] = true
			}
		}
	}
	var chars []*ent.Character
	if includeAll {
		chars, err = db.Client.Character.Query().WithUser().Order(character.ByCampaignID(), character.ByName()).All(ctx)
	} else if len(uidList) == 0 {
		WriteJSON(c, http.StatusOK, []CampaignGroup{})
		return
	} else {
		chars, err = db.Client.Character.Query().Where(character.UserIDIn(uidList...)).WithUser().Order(character.ByCampaignID(), character.ByName()).All(ctx)
	}
	if err != nil {
		WriteError(c, http.StatusInternalServerError, err)
		return
	}
	campaigns := make(map[int64][]PartyMember)
	var uncategorized []PartyMember
	for _, ch := range chars {
		ownerName := ""
		if ch.Edges.User != nil {
			ownerName = ch.Edges.User.Username
		}
		var raceColor string
		if rcMap == nil {
			rcMap = GetRaceColorMap()
		}
		if rc, ok := rcMap[strings.ToLower(strings.TrimSpace(ch.Race))]; ok {
			raceColor = rc
		}
		pm := PartyMember{ID: ch.ID, UserID: ch.UserID, OwnerName: ownerName, Name: ch.Name, Race: ch.Race, RaceColor: raceColor, Class: ch.Class, Level: ch.Level, AC: ch.Ac, HPMax: ch.HpMax, HPCurrent: ch.HpCurrent, TempHP: ch.TempHp, Status: "alive", PortraitURL: ch.PortraitURL, CharacterType: ch.CharacterType, Owned: canEditCharacter(c, ch)}
		if role == "admin" || (ch.CampaignID != 0 && dmCampaignIDs[ch.CampaignID]) {
			pm.DMNotes = ch.DmNotes
		}
		if pm.HPCurrent <= 0 {
			pm.Status = "down"
		} else if float64(pm.HPCurrent)/float64(pm.HPMax) < 0.25 {
			pm.Status = "injured"
		}
		if ch.CampaignID != 0 {
			cid := ch.CampaignID
			pm.CampaignID = &cid
			campaigns[ch.CampaignID] = append(campaigns[ch.CampaignID], pm)
		} else {
			uncategorized = append(uncategorized, pm)
		}
	}
	campNames := make(map[int64]string)
	campPartyNames := make(map[int64]string)
	campOwners := make(map[int64]string)
	for _, ca := range camps {
		campNames[ca.ID] = ca.Name
		campPartyNames[ca.ID] = ca.PartyName
		if ca.Edges.User != nil {
			campOwners[ca.ID] = ca.Edges.User.Username
		}
	}
	groups := make([]CampaignGroup, 0)
	for cid, members := range campaigns {
		groups = append(groups, CampaignGroup{ID: cid, Name: campNames[cid], PartyName: campPartyNames[cid], OwnerName: campOwners[cid], Members: members})
	}
	if len(uncategorized) > 0 {
		groups = append(groups, CampaignGroup{Name: "Uncategorized", Members: uncategorized})
	}
	WriteJSON(c, http.StatusOK, groups)
}

func DoRest(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request.Context()
	var req struct {
		RestType     string `json:"rest_type"`
		HitDiceCount int    `json:"hit_dice_count"`
	}
	if !BindOr400(c, &req) {
		return
	}
	if req.RestType != "short" && req.RestType != "long" {
		WriteError(c, http.StatusBadRequest, strErr("rest_type must be 'short' or 'long'"))
		return
	}
	char, err := db.Client.Character.Get(ctx, charID)
	if err != nil {
		WriteNotFound(c, "character not found")
		return
	}
	if !canEditCharacter(c, char) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	hpHealed := 0
	if req.RestType == "long" {
		hpHealed = char.HpMax - char.HpCurrent
		recoveredHD := max(char.Level/2, 1)
		newHD := min(char.HitDiceCurrent+recoveredHD, char.Level)
		db.Client.Character.UpdateOneID(charID).SetHpCurrent(char.HpMax).SetHitDiceCurrent(newHD).SetDeathSavesSuccesses(0).SetDeathSavesFailures(0).SetConcentratingOn("").Save(ctx)
		if char.ExhaustionLevel > 0 {
			newExhaustion := char.ExhaustionLevel - 1
			db.Client.Character.UpdateOneID(charID).SetExhaustionLevel(newExhaustion).Exec(ctx)
		}
		count, _ := db.Client.CharacterSpellcasting.Query().Where(characterspellcasting.CharacterID(charID)).Count(ctx)
		if count > 0 {
			db.Client.CharacterSpellcasting.Update().Where(characterspellcasting.CharacterID(charID)).SetSlots1Used(0).SetSlots2Used(0).SetSlots3Used(0).SetSlots4Used(0).SetSlots5Used(0).SetSlots6Used(0).SetSlots7Used(0).SetSlots8Used(0).SetSlots9Used(0).Save(ctx)
		} else if char.Class != "" {
			db.Client.CharacterSpellcasting.Create().SetCharacterID(charID).SetAbility("").SetSaveDc(10).SetAttackBonus(0).SetSlots1Used(0).SetSlots2Used(0).SetSlots3Used(0).SetSlots4Used(0).SetSlots5Used(0).SetSlots6Used(0).SetSlots7Used(0).SetSlots8Used(0).SetSlots9Used(0).Save(ctx)
		}
	} else {
		count := max(req.HitDiceCount, 0)
		if count > char.HitDiceCurrent {
			count = char.HitDiceCurrent
		}
		if count == 0 && char.HpMax > 0 {
			count = min(1, char.HitDiceCurrent)
		}
		hitDieSize := 10
		if len(char.HitDice) > 1 {
			dieSizeStr := char.HitDice[2:]
			if d, err2 := strconv.Atoi(dieSizeStr); err2 == nil {
				hitDieSize = d
			}
		}
		conMod := abilityMod(char.Con)
		for i := 0; i < count; i++ {
			result, err := getDicePool().Roll(fmt.Sprintf("1d%d", hitDieSize))
			roll := 1
			if err == nil {
				fmt.Sscanf(string(result.Total), "%d", &roll)
			}
			heal := max(roll+conMod, 1)
			hpHealed += heal
		}
		newHp := min(char.HpCurrent+hpHealed, char.HpMax)
		hpHealed = newHp - char.HpCurrent
		db.Client.Character.UpdateOneID(charID).SetHpCurrent(newHp).SetHitDiceCurrent(char.HitDiceCurrent - count).Save(ctx)
	}
	recoverResourcesForRest(charID, req.RestType)
	db.Client.RestLog.Create().SetCharacterID(charID).SetRestType(req.RestType).SetHpHealed(hpHealed).SetNotes("").Save(ctx)
	SendCharacterUpdate(charID)
	SendPartyUpdate()
	WriteJSON(c, http.StatusOK, gin.H{"ok": true, "hp_healed": hpHealed, "rest_type": req.RestType})
}

func LevelUp(c *gin.Context) {
	charID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	ctx := c.Request.Context()
	char, err := db.Client.Character.Get(ctx, charID)
	if err != nil {
		WriteNotFound(c, "character not found")
		return
	}
	if !canEditCharacter(c, char) {
		WriteError(c, http.StatusForbidden, errAccessDenied)
		return
	}
	newLevel := char.Level + 1
	hitDieSize := 10
	if len(char.HitDice) > 1 {
		dieSizeStr := char.HitDice[2:]
		if d, err2 := strconv.Atoi(dieSizeStr); err2 == nil {
			hitDieSize = d
		}
	}
	conMod := abilityMod(char.Con)
	hpGain := max((hitDieSize/2+1)+conMod, 1)
	newHP := char.HpMax + hpGain
	newCur := min(char.HpCurrent+hpGain, newHP)
	db.Client.Character.UpdateOneID(charID).SetLevel(newLevel).SetHpMax(newHP).SetHpCurrent(newCur).SetHitDiceCurrent(char.HitDiceCurrent + 1).Save(ctx)
	newProf := 2
	if newLevel >= 17 {
		newProf = 6
	} else if newLevel >= 13 {
		newProf = 5
	} else if newLevel >= 9 {
		newProf = 4
	} else if newLevel >= 5 {
		newProf = 3
	}
	if newProf > char.ProficiencyBonus {
		db.Client.Character.UpdateOneID(charID).SetProficiencyBonus(newProf).Save(ctx)
	}
	WriteJSON(c, http.StatusOK, gin.H{"ok": true, "new_level": newLevel, "hp_gain": hpGain, "new_hp_max": newHP})
}
