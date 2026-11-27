# trmnl-pipeline Specification

## Purpose
TBD - created by archiving change add-trmnl-character-roster. Update Purpose after archive.
## Requirements
### Requirement: TRMNL plugin linted on PRs
The CI workflow SHALL lint the TRMNL plugin on every pull request.

#### Scenario: PR lint job
- **WHEN** a pull request is opened or updated
- **THEN** the CI workflow SHALL run `trmnlp lint` against the `trmnl/` plugin project
- **AND** the lint job SHALL fail the PR checks if the plugin files are invalid

#### Scenario: Lint environment
- **WHEN** the lint job runs
- **THEN** it SHALL set up Ruby, install the `trmnl_preview` gem, and run `trmnlp lint` in the `trmnl/` working directory

### Requirement: TRMNL plugin published on release
The release workflow SHALL publish the TRMNL plugin to the TRMNL service after a successful release.

#### Scenario: Push on release
- **WHEN** a release is published on `main`
- **AND** the commit message does not contain `[skip ci]`
- **THEN** the release workflow SHALL run `trmnlp lint` followed by `trmnlp push --force` in the `trmnl/` working directory
- **AND** the push SHALL authenticate with the `TRMNL_API_KEY` environment variable

#### Scenario: Plugin project marker
- **WHEN** tooling inspects the `trmnl/` directory
- **THEN** a `.trmnlp.yml` SHALL be present marking it as a trmnlp plugin project
- **AND** Liquid templates and `settings.yml` SHALL live under `trmnl/src/`
