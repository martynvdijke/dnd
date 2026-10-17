## ADDED Requirements

### Requirement: Outbound media posting from galleries

Villum SHALL send stored uploads to Telegram as photos, audio, video, or documents from the media gallery of a campaign, one-shot, party, item, character, or NPC. A campaign chat target SHALL require the requesting user to be a DM of that campaign and the campaign's Telegram chat to be bound and enabled. A DM target SHALL require the requesting user's Telegram identity to be linked. The requesting user SHALL be able to access every referenced upload under the same ownership rules that already govern upload access elsewhere in Villum.

#### Scenario: Post a photo to a bound campaign chat

- **WHEN** a DM selects a gallery image and posts it to the bound campaign chat
- **THEN** the bot sends it as a photo and reports the per-upload result as sent

#### Scenario: Post media to the user's own DM

- **WHEN** a linked user posts a selection to their own DM
- **THEN** the bot sends the media to that user's chat with the bot

#### Scenario: Non-DM cannot post to a campaign chat

- **WHEN** a non-DM member attempts to post media to a campaign chat
- **THEN** the request is rejected and nothing is sent

#### Scenario: An upload the requester cannot access is rejected

- **WHEN** a user attempts to post an upload that neither they nor a campaign they belong to owns
- **THEN** the request is rejected and no request is made to Telegram

#### Scenario: Post an album

- **WHEN** the user selects several photos or videos and posts them together
- **THEN** the bot sends them as a media group of at most ten items per album batch

#### Scenario: Mixed selection falls back per file

- **WHEN** a selection mixes photos, audio, and documents
- **THEN** photos and videos are grouped as albums, while audio and documents are sent individually

#### Scenario: Unsupported media falls back to a document

- **WHEN** a selected file cannot be sent as a photo, audio, or video (for example PDF, SVG, or WAV)
- **THEN** the bot sends it as a document

### Requirement: Media posting targets

Villum SHALL expose the postable targets for the current user: the campaign chats they can post to (bound and enabled, DM-only) and whether their own DM is linked.

#### Scenario: List postable targets

- **WHEN** the user opens the post-to-Telegram dialog
- **THEN** the response lists the eligible campaign chats and the DM link state

#### Scenario: No targets available

- **WHEN** the user has no bound campaign chat and no linked Telegram identity
- **THEN** the dialog explains that no target is available

### Requirement: Media delivery accounting

Every outbound media send SHALL be recorded with the upload, target type and chat, Telegram message id when accepted, and a status that moves from pending to sent or failed with the last error. A failed send SHALL NOT silently disappear.

#### Scenario: Successful send is recorded

- **WHEN** an upload is sent to a target
- **THEN** a delivery row records the target, status sent, and the Telegram message id

#### Scenario: Failed send reports the error

- **WHEN** Telegram rejects a media send
- **THEN** the response marks that upload failed with an error and the delivery row stores the failure

### Requirement: Inbound media saved from a campaign chat

In a bound and enabled campaign chat, the bot SHALL save media to that campaign's gallery when the message carries an explicit `/save` command or a caption beginning with `/save`. When the campaign's auto-save setting is enabled, the bot SHALL save every media message from that chat. When auto-save is disabled and no `/save` is present, the bot SHALL ignore the media.

#### Scenario: Save with a caption command

- **WHEN** a member posts a photo in the bound chat with the caption "/save"
- **THEN** the photo is stored in the campaign gallery

#### Scenario: Auto-save enabled captures everything

- **WHEN** a member posts media in a campaign chat whose auto-save setting is enabled
- **THEN** the media is stored in the campaign gallery without an explicit command

#### Scenario: Ignored without command and auto-save

- **WHEN** a member posts media in a bound chat with auto-save disabled and no `/save` command
- **THEN** nothing is stored and no error is surfaced

#### Scenario: Oversized inbound media is refused

- **WHEN** an inbound file exceeds Telegram's download limit for the bot
- **THEN** the bot replies that the file is too large and stores nothing

#### Scenario: Duplicate media is skipped

- **WHEN** the same Telegram file arrives twice
- **THEN** the second arrival is ignored and the existing upload is reused

### Requirement: Inbound media from DMs and unbound chats

Media sent to the bot outside a bound campaign chat SHALL be held as a pending capture. The bot SHALL reply with a numbered list of campaigns the sender can save to, and `/save <n>` SHALL save the pending capture to the chosen campaign. When the sender can save to exactly one campaign, the bot SHALL save without the picker. Pending captures SHALL expire after 24 hours and SHALL be removed once consumed.

#### Scenario: DM media prompts a campaign picker

- **WHEN** a linked user sends a photo to the bot in a DM and can access more than one campaign
- **THEN** the bot replies with a numbered campaign list and stores a pending capture

#### Scenario: Save to the chosen campaign

- **WHEN** the user replies "/save 2" after the picker
- **THEN** the pending capture is stored in the second listed campaign's gallery and the pending capture is removed

#### Scenario: Single campaign saves directly

- **WHEN** a linked user with exactly one saveable campaign sends media in a DM
- **THEN** the media is stored in that campaign without a picker prompt

#### Scenario: Expired pending capture

- **WHEN** a user replies "/save" for a pending capture older than 24 hours
- **THEN** the bot reports that the capture expired and stores nothing

### Requirement: Media sending limits

Outbound sends SHALL respect Telegram limits by pre-validating file sizes and sizes: photos up to 10 MB, other media up to 50 MB per file, and albums of at most ten items. Captions SHALL be truncated to Telegram's 1024-character limit. A file over the limit SHALL be reported as failed before upload.

#### Scenario: Oversized video is rejected before sending

- **WHEN** the user posts a video larger than the per-file send limit
- **THEN** the response marks it failed with a size error and no request is made to Telegram

#### Scenario: Long caption is truncated

- **WHEN** a derived caption exceeds the limit
- **THEN** it is truncated to the allowed length before sending

### Requirement: Media format selection

Uploads SHALL be sent using the Telegram method matching their kind where Telegram supports the format (JPEG/PNG/WebP/GIF as photos, MP3/OGG/Opus as audio, MP4/WebM/MOV as video) and as documents otherwise.

#### Scenario: MP3 is sent as audio

- **WHEN** the user posts an MP3 upload
- **THEN** the bot sends it with the audio method

#### Scenario: WAV is sent as a document

- **WHEN** the user posts a WAV upload
- **THEN** the bot sends it as a document

#### Scenario: PDF is sent as a document

- **WHEN** the user posts a PDF upload
- **THEN** the bot sends it as a document
