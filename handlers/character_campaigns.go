package handlers

import (
	"context"
	"sort"

	"villum/db"
	"villum/ent/campaigncharacter"
	"villum/ent/character"
	"villum/ent/predicate"
	"villum/models"
)

// characterInCampaign matches characters that have a membership in the campaign.
func characterInCampaign(campaignID int64) predicate.Character {
	return character.HasCampaignLinksWith(campaigncharacter.CampaignID(campaignID))
}

// characterCampaignRefs loads the campaigns each character belongs to, keyed by
// character ID. Each list is ordered by campaign name.
func characterCampaignRefs(ctx context.Context, characterIDs []int64) (map[int64][]models.CampaignRef, error) {
	out := make(map[int64][]models.CampaignRef, len(characterIDs))
	if len(characterIDs) == 0 {
		return out, nil
	}
	rows, err := db.Client.CampaignCharacter.Query().
		Where(campaigncharacter.CharacterIDIn(characterIDs...)).
		WithCampaign().
		All(ctx)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		name := ""
		if row.Edges.Campaign != nil {
			name = row.Edges.Campaign.Name
		}
		out[row.CharacterID] = append(out[row.CharacterID], models.CampaignRef{ID: row.CampaignID, Name: name})
	}
	for id, refs := range out {
		sort.Slice(refs, func(i, j int) bool { return refs[i].Name < refs[j].Name })
		out[id] = refs
	}
	return out, nil
}

// characterCampaigns returns the campaigns a single character belongs to.
func characterCampaigns(ctx context.Context, characterID int64) []models.CampaignRef {
	refs, err := characterCampaignRefs(ctx, []int64{characterID})
	if err != nil {
		return []models.CampaignRef{}
	}
	if refs[characterID] == nil {
		return []models.CampaignRef{}
	}
	return refs[characterID]
}

// characterMemberCampaignIDs returns the campaign IDs a character belongs to.
func characterMemberCampaignIDs(ctx context.Context, characterID int64) ([]int64, error) {
	rows, err := db.Client.CampaignCharacter.Query().
		Where(campaigncharacter.CharacterID(characterID)).
		Select(campaigncharacter.FieldCampaignID).
		All(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.CampaignID)
	}
	return ids, nil
}

// replaceCharacterCampaigns reconciles a character's campaign memberships to
// exactly the given set. The caller must already have validated permissions.
func replaceCharacterCampaigns(ctx context.Context, characterID int64, campaignIDs []int64) error {
	current, err := characterMemberCampaignIDs(ctx, characterID)
	if err != nil {
		return err
	}
	want := make(map[int64]bool, len(campaignIDs))
	for _, id := range campaignIDs {
		want[id] = true
	}
	have := make(map[int64]bool, len(current))
	for _, id := range current {
		have[id] = true
	}
	for _, id := range current {
		if !want[id] {
			if _, err := db.Client.CampaignCharacter.Delete().
				Where(campaigncharacter.CharacterID(characterID), campaigncharacter.CampaignID(id)).
				Exec(ctx); err != nil {
				return err
			}
		}
	}
	for _, id := range campaignIDs {
		if id == 0 || have[id] {
			continue
		}
		if err := db.Client.CampaignCharacter.Create().
			SetCharacterID(characterID).
			SetCampaignID(id).
			Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
