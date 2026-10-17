## ADDED Requirements

### Requirement: Audio and video upload support

The upload endpoint SHALL accept audio files (`.mp3 .ogg .oga .wav .m4a .aac .flac .opus`) up to 25 MB and video files (`.mp4 .webm .mov .m4v`) up to 50 MB, in addition to the existing image and PDF support. The declared extension and the detected content type SHALL agree with the file's media kind; mismatches SHALL be rejected.

#### Scenario: Upload an MP3

- **WHEN** a user uploads an MP3 under the audio size limit
- **THEN** the upload is stored with kind audio and its URL is returned

#### Scenario: Upload an MP4

- **WHEN** a user uploads an MP4 under the video size limit
- **THEN** the upload is stored with kind video and its URL is returned

#### Scenario: Audio over the size limit is rejected

- **WHEN** a user uploads an audio file larger than the audio limit
- **THEN** the upload is rejected with a size error

#### Scenario: Content type mismatch is rejected

- **WHEN** a file's detected content type does not match its media kind
- **THEN** the upload is rejected

#### Scenario: Image and PDF behavior is unchanged

- **WHEN** a user uploads an image or PDF within the existing limits
- **THEN** the upload behaves exactly as before, including resizing and thumbnails for images

### Requirement: Upload media metadata

Stored uploads SHALL carry a media kind (image, pdf, audio, video, or other), a source (upload or telegram), and the Telegram file unique id when the file arrived from Telegram. Existing uploads SHALL be backfilled with an appropriate kind.

#### Scenario: Uploaded file carries its kind and source

- **WHEN** a file is uploaded through the web UI
- **THEN** the stored upload records kind from its extension and source upload

#### Scenario: Telegram file carries its identifiers

- **WHEN** a file is saved from Telegram
- **THEN** the stored upload records source telegram and the Telegram file unique id

### Requirement: Gallery media playback

Media galleries SHALL render images as image previews, PDFs as a document link or icon, audio with an audio player, and video with a video player. Gallery tiles SHALL keep their copy and delete actions for every kind.

#### Scenario: Audio plays in the gallery

- **WHEN** a gallery contains an audio upload
- **THEN** the tile renders an audio player that can play the file

#### Scenario: Video plays in the gallery

- **WHEN** a gallery contains a video upload
- **THEN** the tile renders a video player that can play the file

#### Scenario: PDF still renders as a document

- **WHEN** a gallery contains a PDF upload
- **THEN** the tile renders the document affordance rather than a media player

### Requirement: File picker accepts audio and video

The file picker SHALL offer audio and video files for selection in addition to images, and the modal upload path SHALL submit files under the field name expected by the upload endpoint.

#### Scenario: Pick an audio file

- **WHEN** the user opens the file picker
- **THEN** audio and video files can be selected and uploaded

#### Scenario: Modal upload succeeds

- **WHEN** a file is submitted through the upload modal
- **THEN** the server stores it rather than rejecting the missing file field
