# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

### Added
- Added regression coverage for channel cache refresh after channel create, update, and delete operations.

### Fixed
- Refreshed the in-memory channel cache immediately after channel create, update, delete, and status changes so newly configured channels are available without waiting for the periodic sync loop.
- Invalidated cached per-group model lists after channel changes so `/v1/models` and request routing stay consistent after channel edits.

### Docs
- Documented the channel cache refresh behavior in the Chinese and English FAQ sections.
