## Why

Villum already delivers text recaps to Telegram, but the rich media DMs actually use — handout images, battle maps, audio stings, PDFs, short video clips — lives only in Villum galleries and cannot reach the table's chats; media players send into Telegram never reaches Villum either. Uploads also reject audio and video outright. Closing that gap makes the bot useful for the whole table's media, not just recaps.

## What Changes

- Extend uploads with audio (`.mp3 .ogg .oga .wav .m4a .aac .flac .opus`) and video (`.mp4 .webm .mov .m4v`): per-kind size caps (image/PDF 10 MB unchanged, audio 25 MB, video 50 MB), kind-aware MIME validation, and new `uploads` metadata: `kind`, `source` (`upload` | `telegram`), `telegram_file_unique_id`.
- Render audio and video in media galleries (`<audio controls>`, `<video controls>`) and accept audio/video in the file picker; fix the `ts/app/uploads.ts` form-field mismatch (`file` → `image`).
- Extend the Telegram client with binary multipart senders (`sendPhoto`, `sendAudio`, `sendVideo`, `sendDocument`) and `sendMediaGroup` albums, plus `getFile`/file download for inbound media; parse photo/document/audio/video/voice/caption/media-group fields from updates.
- **Outbound:** `POST /api/telegram/media` posts one or many uploads to a bound campaign chat or the caller's DM, with an optional recap reference for captions and delivery accounting; `GET /api/telegram/targets` lists postable chats; sends are recorded in a new `telegram_media_posts` table.
- **Inbound:** media posted in a bound campaign group is saved to that campaign's gallery on explicit `/save` (caption or command); a per-campaign `auto_save_media` toggle (default off) saves all group media. Media in DMs or unbound chats is held as a pending capture and the bot replies with a numbered campaign picker consumed by `/save <n>`; duplicates are skipped by `telegram_file_unique_id`; inbound files are capped at Telegram's 20 MB download limit.
- Recaps can optionally carry selected media: `media_upload_ids` on recap create/mark-sent sends the selected uploads to the same targets after the recap text, recorded against the recap.
- Gallery UI: per-item "Post to Telegram" plus multi-select album posting with a target dialog (campaign chats + own DM) and optional recap attachment.
- Extend the Telegram test mock with media methods, multipart parsing and file downloads; add Go/e2e coverage for both directions.

## Capabilities

### New Capabilities
- `telegram-media`: outbound photo/audio/video/document posting and albums, target discovery, delivery accounting, inbound media capture with `/save` and campaign picker, per-campaign auto-save, and duplicate/size handling.
- `media-uploads`: audio/video upload support (extensions, size caps, MIME/kind validation), upload metadata (kind/source/telegram id), gallery playback, and picker acceptance.

### Modified Capabilities
- `telegram-recap-bot`: recap delivery can include selected media attachments (new requirement; existing delivery semantics unchanged).

## Impact

- Backend: `telegram/client.go`, `telegram/dispatcher.go`, `telegram/handlers.go`, `telegram/routing.go`, `telegram/store.go`, `telegram/config.go`, new `telegram/media.go`, `handlers/upload.go`, `handlers/routes_telegram.go`, `handlers/recaps.go`, `handlers/htmx.go` (media gallery data), `app.go` wiring (media saver callback).
- Schema: new migration adding `uploads.kind/source/telegram_file_unique_id`, `campaign_telegram_settings.auto_save_media`, `telegram_media_posts`, `telegram_pending_captures`.
- Frontend: `handlers/templates/media_gallery.html`, `media_gallery_item.html`, recaps UI, `ts/file-picker.ts`, `ts/app/uploads.ts`, new `ts/telegram-media.ts`; admin/campaign Telegram sections gain the auto-save toggle.
- Tests: `telegram/telegram_test.go`, `handlers/telegram_extra_test.go`, upload tests, `tests/telegram-mock.ts`, `tests/telegram.spec.ts`, upload e2e; coverage floors stay enforced.
- No breaking API changes: new fields are optional and existing recap delivery keeps working unchanged.
