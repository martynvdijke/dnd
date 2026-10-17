## Context

The recap bot (migration 062, `telegram/` package) can send text and plain-text documents to bound campaign chats and opted-in DMs. Everything media-related stops at text: the client has no binary multipart senders, `Message` carries no media fields, the dispatcher drops updates without text, and uploads only accept images and PDFs. Galleries exist for campaigns and one-shots, and they sit behind authentication on local disk (`MediaPath`, default `/app/media`), so Telegram can never fetch them by URL — bytes must be pushed from the server.

Constraints: Telegram Bot API file limits (photos ≤10 MB, other uploads ≤50 MB, `getFile` downloads ≤20 MB); existing recap delivery must keep working; the `telegram` package cannot import `handlers` (import cycle) so the media save path needs a callback seam; all mutations need CSRF/API-token handling already used by the SPA.

## Goals / Non-Goals

**Goals:**

- Upload, store, render and play audio and video in Villum, with per-kind size/MIME validation.
- Push photos, audio, video, documents and albums from Villum galleries to a campaign chat or the caller's DM, with optional recap attachment.
- Capture media posted in a bound campaign chat into that campaign's gallery on explicit `/save` or (optionally) automatically; route DMs/unbound chats through a numbered campaign picker.
- Keep delivery idempotent and observable (`telegram_media_posts`, per-upload statuses), and avoid duplicates via Telegram's `file_unique_id`.

**Non-Goals:**

- Streaming/bot-side transcoding, thumbnails from Telegram, sticker/animation/poll handling.
- Editing or deleting Telegram messages after send; retrying failed media posts beyond the row accounting.
- Media on share links, one-off inline bots, or Telegram as a storage backend.
- Changing recap text delivery semantics.

## Decisions

1. **Server pushes bytes via multipart; never public URLs.** `telegram/client.go` gains `SendMedia(chatID, kind, filename, data []byte)` plus `SendMediaGroup(chatID, []MediaItem)` using multipart fields `attach://file0…` for albums ≤10. Rationale: media files are auth-gated and local; public URLs are impossible without exposing media. Fallback mapping: `.jpg/.jpeg/.png/.webp/.gif` → `sendPhoto`; `.mp3/.ogg/.oga/.opus` → `sendAudio`; `.mp4/.webm/.mov` → `sendVideo`; anything else (SVG, PDF, wav, m4a, flac, avif) → `sendDocument`. Missing/no-extension files from inbound are sent as documents.
   - Alternative: public signed URLs — rejected (would expose a media proxy surface, and Telegram caches files anyway).

2. **`saveMediaBytes` seam for both directions.** Extract the existing upload pipeline (sha256 dedupe, disk write, resize/thumbnail, `uploads` INSERT) into `saveMediaBytes(name string, data []byte, ownerType string, ownerID *int64, source, telegramFileUniqueID string) (*models.Upload, bool, error)` in `handlers/upload.go`; `HandleUpload` calls it, and `telegram` gets `SetMediaSaver(func(MediaSaveRequest) (*MediaSaveResult, error))` wired in `app.go`. Rationale: no import cycle, one dedupe/resize path; inbound dedupe by `telegram_file_unique_id` plus hash.

3. **Kind is first-class metadata.** Migration 067 adds `uploads.kind TEXT DEFAULT 'image'`, `uploads.source TEXT DEFAULT 'upload'`, `uploads.telegram_file_unique_id TEXT`, `campaign_telegram_settings.auto_save_media INTEGER DEFAULT 0`, plus tables `telegram_media_posts` and `telegram_pending_captures`. Existing rows are backfilled `kind='pdf'` for `ext IN ('pdf')` else `'image'`. Detection is by extension first, with MIME cross-check (`kindFromExt` + `detectMediaKind`), shared by upload and inbound. Size caps: image/PDF 10 MB (unchanged), audio 25 MB, video 50 MB; inbound additionally capped at 20 MB (Bot API download limit).
   - Alternative: separate `media` table — rejected; `uploads` already has the gallery/link/dedupe machinery.

4. **Outbound API shape.** `POST /api/telegram/media` body `{upload_ids:[…], target:{kind:"campaign"|"dm", campaign_id?:N}, recap_id?:N, caption?:string}` → per-upload results `{upload_id, status:"sent"|"failed", telegram_message_id?, error?}`. `GET /api/telegram/targets` → `{campaigns:[{id,title,chat_title}], dm:{linked:bool}}`. Campaign targets require `isCampaignDM` and `is_enabled=1`; DM target requires a linked identity. Albums: consecutive photo/video uploads batched ≤10 via `sendMediaGroup`; audio/documents sent individually (per-file captions). Every send writes a `telegram_media_posts` row (status pending→sent/failed, `recap_id` when attached); captions default to the upload filename, or the recap title when attached.
   - Alternative: reuse `telegram_deliveries` — rejected; that table is recap-specific (`UNIQUE(recap_id,target_chat_id)`).

5. **Inbound capture is explicit by default.** In a bound, enabled campaign chat:
   - media with `/save` command or caption starting `/save` → saved to the campaign gallery immediately;
   - when the campaign's `auto_save_media=1` → every media message saved;
   - otherwise ignored.
   In a DM or unbound chat: media is stored as a `telegram_pending_captures` row (`telegram_user_id`, `chat_id`, `file_unique_id`, `file_id`, `kind`, `file_name`, `caption`, `created_at`) and the bot replies with the user's campaigns as a numbered list; `/save <n>` consumes the most recent pending capture into the chosen campaign (allowed for owned/dmed campaigns only, reusing `memberCampaignIDs`). A single-campaign user saves without the picker. Pending rows expire after 24 h and are deleted on use; duplicates are skipped via `telegram_file_unique_id UNIQUE` on uploads.
   - Alternative: callback buttons — deferred; text picker needs no new update types and works in both poller and webhook modes.

6. **Recap media attachments ride the existing delivery path.** `CreateCampaignRecap` and `MarkRecapAsSent` accept optional `media_upload_ids`; after text delivery finishes, `DeliverRecap` also sends those uploads to the same targets and records them in `telegram_media_posts` with `recap_id`. Media failures never flip the recap delivery itself to failed.
   - Alternative: a separate "attach" endpoint after sending — more clicks, no atomic recap reference.

7. **Dispatcher rework.** `Message` gains `Caption`, `Photo []PhotoSize`, `Document`, `Audio`, `Video`, `Voice`, `MediaGroupID`. `HandleUpdate` keeps command handling but delegates non-text messages to `handleInboundMedia`, which resolves the chat's binding (`campaign_telegram_settings` by `chat_id`) and applies decision 5. Chat-type is stored at bind time already; commands in groups keep working.
   - Alternative: update-type dispatch table — unnecessary today (one update kind).

8. **Gallery UI.** `media_gallery.html` gains per-item checkboxes, a "Post to Telegram" button, and an item action; a `ts/telegram-media.ts` module fetches `/api/telegram/targets`, shows the target dialog (campaign select or DM), validates sizes, posts, and surfaces per-file results. `media_gallery_item.html` renders `<audio controls>`/`<video controls>`; file picker and upload accept lists gain audio/video; `ts/app/uploads.ts` field bug fixed (`file`→`image`).
   - Alternative: separate media-posting page — rejected; posting belongs next to the files.

9. **Test seam.** `tests/telegram-mock.ts` gains multipart parsing for `sendPhoto/sendAudio/sendVideo/sendDocument/sendMediaGroup` (bytes discarded, metadata captured), `getFile` returning a file path and `GET /file/bot<token>/<path>` serving fixture bytes. E2E covers: bind chat → gallery post → mock receives photo/album; webhook media update with `/save` → upload appears in gallery; auto-save toggle; DM picker flow.

## Risks / Trade-offs

- **Telegram size limits reject large files.** Mitigate with pre-flight caps and clear per-file errors; 50 MB video cap means big uploads can still fail and are reported, not silently dropped.
- **Inbound media doubles storage** for auto-save campaigns; mitigate with per-campaign opt-in (default off) and duplicate skip.
- **Multipart `sendMediaGroup` is fiddly** (`attach://` + repeated `media` JSON). Mitigate with a focused client unit test against the mock server.
- **Pending-capture picker can mis-target** if the user has many campaigns; mitigate by echoing the chosen campaign in the confirmation reply and deleting the pending row on use.
- **Kind backfill for existing rows**: old violet rows keep working since `kind` defaults to `image` and only PDFs differ; gallery rendering decides by `kind`/`ext`, not `source`.
- **Mock/real Telegram divergence** (multipart shapes, file download path) is the main e2e blind spot; keep client unit tests asserting exact multipart field names.
