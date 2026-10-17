## Context

`share_links` (token, entity_type, entity_id, created_by, expires_at) already backs public sharing for character, party, note, journal, map, and upload. `GetSharedEntity` serves JSON for all types; `GetSharedPage` serves server-rendered HTML for note/journal/map/upload. Recaps live in `campaign_recaps` (campaign_id, title, content, session dates, word_count) and are gated by `isCampaignMember`/`isCampaignDM`.

## Goals / Non-Goals

**Goals**
- Let a campaign member publish a recap as a revocable public link.
- Reuse the existing token/expiry/revocation machinery unchanged.
- Render a standalone, noindex HTML page consistent with the other share pages.

**Non-Goals**
- No new table or column; no per-recap `is_public` flag.
- No public listing of recaps; links are unguessable tokens only.
- No editing of recaps from the public page.

## Decisions

- **Reuse `share_links` rather than adding `is_public`/`token` to `campaign_recaps`.** The generic link table already handles expiry, revocation, and listing; a per-recap flag would duplicate that and lose revocation.
- **Ownership = campaign membership.** `canShareEntity("recap")` resolves the recap's `campaign_id` and delegates to `userCanAccessCampaign`, matching the `party` case. Any member (not just DM) may share, consistent with read access.
- **HTML page uses `html/template` autoescaping** via the existing `renderSharePage`, so recap content is escaped. Content is stored as HTML from the TipTap editor; it is rendered inside a `.content` div (same trust model as the note share page).

## Risks / Trade-offs

- Recap content is DM-authored HTML; the public page renders it as HTML. This matches the existing note/journal share behavior and the content is only reachable via an unguessable token.
- No expiry by default (same as other share types); users can set `expires_in` and revoke from the Shared Links dialog.
