# localization Specification

## Purpose
TBD - created by archiving change add-i18n-localization. Update Purpose after archive.
## Requirements
### Requirement: Runtime UI translation
The frontend SHALL translate user-facing strings at runtime through a `t(key, params?)` function backed by per-locale dictionaries, without adding a third-party i18n dependency.

#### Scenario: Known key translated
- **WHEN** the active locale is `nl` and `t('nav.characters')` is called
- **THEN** the Dutch translation for that key is returned

#### Scenario: Placeholder interpolation
- **WHEN** a dictionary value contains `{name}` and `t(key, {name: 'Aria'})` is called
- **THEN** the returned string has `{name}` replaced with `Aria`

#### Scenario: Unknown key falls back
- **WHEN** `t('does.not.exist')` is called
- **THEN** the key itself is returned and no error is thrown

### Requirement: Attribute-driven translation of static markup
The frontend SHALL translate elements carrying `data-i18n`, `data-i18n-placeholder`, `data-i18n-title`, or `data-i18n-aria-label` attributes when `applyTranslations(root)` runs.

#### Scenario: Text content translated
- **WHEN** `applyTranslations` runs over an element with `data-i18n="nav.dice"`
- **THEN** the element's text content is set to the active locale's translation

#### Scenario: Attribute translated
- **WHEN** `applyTranslations` runs over an input with `data-i18n-placeholder="search.placeholder"`
- **THEN** the input's `placeholder` attribute is set to the active locale's translation

### Requirement: Locale selection and persistence
The frontend SHALL detect the initial locale from a stored preference, then the browser language, then English, and SHALL persist an explicit user selection.

#### Scenario: Stored preference wins
- **WHEN** `villum.locale` is `nl` in localStorage
- **THEN** the initial locale is `nl`

#### Scenario: Browser language used when unset
- **WHEN** no stored preference exists and `navigator.language` is `nl-NL`
- **THEN** the initial locale is `nl`

#### Scenario: Selection persisted
- **WHEN** the user selects a locale from the language switcher
- **THEN** the choice is stored and translations are re-applied immediately

### Requirement: Language switcher
The app navigation SHALL expose a language switcher listing the available locales.

#### Scenario: Switcher changes language
- **WHEN** the user picks a different locale in the switcher
- **THEN** the navigation labels update to that locale without a page reload
