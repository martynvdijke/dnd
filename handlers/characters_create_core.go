package handlers

import (
	"context"
	"strings"
	"time"

	"villum/db"
	"villum/models"
	"villum/telegram"
)

// statusError carries an HTTP status for errors raised while creating a
// character so the web handler can answer with the right code.
type statusError struct {
	Status  int
	Message string
}

func (e *statusError) Error() string { return e.Message }

// createCharacterCore validates and persists a character owned by uid. The web
// handler and the Telegram bot both use it so validation, defaults, currency,
// and campaign membership behave identically on every creation path.
func createCharacterCore(ctx context.Context, uid int64, ch models.Character) (models.Character, error) {
	if strings.TrimSpace(ch.Name) == "" {
		return ch, &statusError{Status: 400, Message: "name is required"}
	}
	if ch.CharacterType == "" {
		ch.CharacterType = "player"
	}
	if ch.CharacterType != "player" && ch.CharacterType != "linked" {
		return ch, &statusError{Status: 400, Message: "character_type must be 'player' or 'linked'"}
	}
	for _, score := range []int{ch.Str, ch.Dex, ch.Con, ch.Int, ch.Wis, ch.Cha} {
		if score < 0 || score > 30 {
			return ch, &statusError{Status: 400, Message: "ability scores must be between 0 and 30"}
		}
	}

	// Set defaults
	if ch.Level < 1 {
		ch.Level = 1
	}
	if ch.AC < 1 {
		ch.AC = 10
	}
	if ch.Speed < 1 {
		ch.Speed = 30
	}
	if ch.ProficiencyBonus < 1 {
		if ch.Level >= 17 {
			ch.ProficiencyBonus = 6
		} else if ch.Level >= 13 {
			ch.ProficiencyBonus = 5
		} else if ch.Level >= 9 {
			ch.ProficiencyBonus = 4
		} else if ch.Level >= 5 {
			ch.ProficiencyBonus = 3
		} else {
			ch.ProficiencyBonus = 2
		}
	}
	if ch.HPMax < 1 {
		ch.HPMax = 10
		ch.HPCurrent = 10
	}
	if ch.HitDice == "" {
		ch.HitDice = "1d10"
		ch.HitDiceCurrent = 1
	}

	now := time.Now().Format("2006-01-02 15:04:05")

	charCreate := db.Client.Character.Create().
		SetUserID(uid).
		SetName(ch.Name).
		SetRace(ch.Race).
		SetClass(ch.Class).
		SetSubclass(ch.Subclass).
		SetLevel(ch.Level).
		SetXp(ch.XP).
		SetBackground(ch.Background).
		SetAlignment(ch.Alignment).
		SetStr(ch.Str).
		SetDex(ch.Dex).
		SetCon(ch.Con).
		SetInt(ch.Int).
		SetWis(ch.Wis).
		SetCha(ch.Cha).
		SetAc(ch.AC).
		SetInitiative(ch.Initiative).
		SetSpeed(ch.Speed).
		SetHpMax(ch.HPMax).
		SetHpCurrent(ch.HPCurrent).
		SetTempHp(ch.TempHP).
		SetHitDice(ch.HitDice).
		SetHitDiceCurrent(ch.HitDiceCurrent).
		SetProficiencyBonus(ch.ProficiencyBonus).
		SetInspiration(ch.Inspiration).
		SetPassivePerception(ch.PassivePerception).
		SetDeathSavesSuccesses(ch.DeathSavesSuccesses).
		SetDeathSavesFailures(ch.DeathSavesFailures).
		SetConcentratingOn(ch.ConcentratingOn).
		SetPersonalityTraits(ch.PersonalityTraits).
		SetIdeals(ch.Ideals).
		SetBonds(ch.Bonds).
		SetFlaws(ch.Flaws).
		SetAppearance(ch.Appearance).
		SetBackstory(ch.Backstory).
		SetPortraitURL(ch.PortraitURL).
		SetCreatedAt(now).
		SetUpdatedAt(now).
		SetCharacterType(ch.CharacterType)

	desiredCampaigns := make([]int64, 0, len(ch.CampaignIDs))
	for _, cid := range ch.CampaignIDs {
		if cid == 0 {
			continue
		}
		member, err := IsCampaignMember(ctx, cid, uid)
		if err != nil {
			return ch, err
		}
		if !member {
			return ch, &statusError{Status: 403, Message: "cannot add character to a campaign you are not a member of"}
		}
		desiredCampaigns = append(desiredCampaigns, cid)
	}

	char, err := charCreate.Save(ctx)
	if err != nil {
		return ch, err
	}

	id := char.ID

	// Create default currency entry
	db.Client.CharacterCurrency.Create().SetCharacterID(id).Save(ctx)

	ch.ID = id
	ch.UserID = uid
	if err := replaceCharacterCampaigns(ctx, id, desiredCampaigns); err != nil {
		return ch, err
	}
	ch.Campaigns = characterCampaigns(ctx, id)
	db.DB.Exec("PRAGMA wal_checkpoint(PASSIVE)")
	return ch, nil
}

// BotCharacterCreator adapts createCharacterCore for the Telegram /create flow.
func BotCharacterCreator(ctx context.Context, userID int64, in telegram.CreateCharacterInput) (telegram.CreatedCharacter, error) {
	ch := models.Character{
		Name:          in.Name,
		Race:          in.Race,
		Class:         in.Class,
		Level:         in.Level,
		CharacterType: "player",
	}
	if in.CampaignID != 0 {
		ch.CampaignIDs = []int64{in.CampaignID}
	}
	created, err := createCharacterCore(ctx, userID, ch)
	if err != nil {
		return telegram.CreatedCharacter{}, err
	}
	return telegram.CreatedCharacter{
		ID:        created.ID,
		Name:      created.Name,
		Race:      created.Race,
		Class:     created.Class,
		Level:     created.Level,
		HpMax:     created.HPMax,
		HpCurrent: created.HPCurrent,
		Ac:        created.AC,
	}, nil
}
