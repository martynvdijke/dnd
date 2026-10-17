## ADDED Requirements

### Requirement: Recap media attachments

When a recap is sent to Telegram, Villum SHALL optionally send selected media uploads to the same targets after the recap text. Attachments SHALL be recorded against the recap in the media delivery accounting. A media attachment failure SHALL NOT change the recap text delivery status.

#### Scenario: Send a recap with attachments

- **WHEN** a recap is sent with selected media uploads
- **THEN** the recap text is delivered and the selected media follows to the same targets, recorded against the recap

#### Scenario: Attachment failure does not fail the recap

- **WHEN** one attached upload fails to send
- **THEN** the recap text delivery status is unaffected and the failure is reported for that upload
