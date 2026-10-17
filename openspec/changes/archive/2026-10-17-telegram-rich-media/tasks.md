## 1. Schema and upload foundation

- [x] 1.1 Migration 067: `campaign_telegram_settings.auto_save_media` (default 0) and tables `telegram_media_posts`, `telegram_pending_captures`; register in `db/migrations/registry.go`. The `uploads.kind/source/telegram_file_unique_id` columns and their indexes + `kind='pdf'` backfill live in `db/migrations/safe_alters.go` instead, because `uploads` is ent-owned and `ent.Schema.Create` rebuilds the table after `Migrate()`, dropping columns added earlier in the chain
- [x] 1.2 Extend `handlers/upload.go`: audio (`.mp3 .ogg .oga .wav .m4a .aac .flac .opus`) and video (`.mp4 .webm .mov .m4v`) extensions, kind detection (extension first, MIME cross-check), kind-aware caps (image/PDF 10 MB, audio 25 MB, video 50 MB), `kind` in the JSON response and `models.Upload`
- [x] 1.3 Extract `saveMediaBytes(name, data, ownerType, ownerID *int64, source, telegramFileUniqueID)` from `HandleUpload` (hashing, dedupe, resize/thumbnail, insert) and refactor `HandleUpload` to use it
- [x] 1.4 Add `telegram.SetMediaSaver` callback and wire it in `app.go` to the handler save helper

## 2. Telegram client media support

- [x] 2.1 Parse photo/document/audio/video/voice/caption/media_group_id in `Update`/`Message`
- [x] 2.2 Add binary multipart senders `SendPhoto`, `SendAudio`, `SendVideo`, `SendDocument` and a `SendMedia` dispatcher implementing the format-fallback mapping (jpg/jpeg/png/webp/gif → photo, mp3/ogg/oga/opus → audio, mp4/webm/mov → video, else document)
- [x] 2.3 Add `SendMediaGroup` (≤10 items, `attach://fileN` fields) and `GetFile`/file download for inbound media
- [x] 2.4 Client unit tests: multipart field names, format fallback, caption truncation to 1024, group batching, getFile URL, RetryAfter handling

## 3. Outbound posting API

- [x] 3.1 `telegram_media_posts` store operations and media service (load upload, choose send method, record pending/sent/failed with `telegram_message_id`, `recap_id` when supplied)
- [x] 3.2 `GET /api/telegram/targets` returning bound campaign chats for the caller's campaigns plus the caller's DM link state
- [x] 3.3 `POST /api/telegram/media` accepting `{upload_ids, target:{kind:'campaign'|'dm',campaign_id?}, recap_id?, caption?}`: campaign target requires DM role and `is_enabled=1`, DM target requires a linked identity, per-upload result array, size pre-validation before sending
- [x] 3.4 Handler tests: campaign post recorded, own-DM post, non-DM rejected, oversized pre-reject, per-file failure reporting

## 4. Inbound media capture

- [x] 4.1 Expose `auto_save_media` on `GET/PUT /api/telegram/campaigns/:id/telegram` (DM-only), default off
- [x] 4.2 Dispatcher inbound media path: explicit `/save` command or caption prefix in a bound enabled chat saves to that campaign's gallery (caption as name/caption); `auto_save_media=1` saves all group media; unsaved group media without a save request is ignored
- [x] 4.3 DM/unbound handling: store a `telegram_pending_captures` row, reply with a numbered campaign picker; `/save <n>` consumes the most recent pending capture; a single-campaign user saves directly; pending captures expire after 24 h
- [x] 4.4 Enforce the 20 MB inbound download cap with a friendly reply, and skip duplicates by `telegram_file_unique_id`
- [x] 4.5 Handler/dispatcher tests for each inbound path (save, auto-save, ignore, oversized, duplicate, DM picker, expiry)

## 5. Recap media attachments

- [x] 5.1 Accept optional `media_upload_ids` on recap create and mark-sent, persisting them for the recap
- [x] 5.2 `DeliverRecap` sends selected media after the recap text to each target and records `telegram_media_posts` rows with `recap_id`; attachment failures are reported per file without changing recap text delivery status
- [x] 5.3 Tests: recap with attachments sends and records; attachment failure leaves recap delivery status intact

## 6. Gallery and TypeScript UI

- [x] 6.1 `media_gallery.html`/`media_gallery_item.html`: render audio (`<audio controls>`) and video (`<video controls>`), per-item "Post to Telegram" button, multi-select checkboxes with album posting, and a target dialog (campaign chats + own DM, optional recap), with data-testids
- [x] 6.2 New `ts/telegram-media.ts` (targets fetch, post dialog, album batching) and campaign Telegram section `auto_save_media` toggle
- [x] 6.3 `ts/file-picker.ts` accepts audio/video; fix `ts/app/uploads.ts` FormData field (`file` → `image`) so modal uploads work
- [x] 6.4 Unit tests for `ts/telegram-media.ts` and the picker/modal changes; keep vitest coverage floors

## 7. Verification and delivery

- [x] 7.1 Extend `tests/telegram-mock.ts`: multipart parsing for photo/audio/video/document, `sendMediaGroup`, `getFile`, and `GET /file/bot<token>/<path>` serving fixture bytes
- [x] 7.2 Playwright: post a gallery image to a bound campaign chat and assert the mock received it; album posting; inbound `/save` via webhook saves to the gallery; auto-save toggle; DM picker flow; upload an audio file end-to-end
- [x] 7.3 `openspec validate telegram-rich-media`; `task ci` green (Go + vitest + e2e + coverage gates); PR opened and merged per the workflow contract
- [x] 7.4 Post-merge: archive the change in a follow-up PR (`openspec archive telegram-rich-media`, `git add -f`), matching the repo convention
