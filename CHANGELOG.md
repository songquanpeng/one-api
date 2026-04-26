# Changelog

All notable changes to this project will be documented in this file.

## Unreleased

### Added
- Added a dedicated `AliBailianCodingPlan` channel type for Alibaba Bailian Coding Plan chat completions.
- Added OpenAI-compatible model presets for Alibaba Bailian and Coding Plan channels.
- Added DeepSeek V4 and refreshed Alibaba Bailian Qwen, DeepSeek, Kimi, GLM, MiniMax, and embedding model presets from the latest provider documentation.
- Added regression coverage for Bailian compatible-mode Base URL handling, DeepSeek thinking passthrough, and updated model billing ratios.
- Added regression coverage for Coding Plan request URL routing and compatible metadata exposure.
- Added regression coverage for channel cache refresh after channel create, update, and delete operations.

### Changed
- Exposed Alibaba Bailian and Coding Plan channel options in the `default`, `air`, and `berry` themes.
- Improved channel form prompts so Bailian and Coding Plan keys are clearly distinguished.
- Updated the Alibaba Bailian default Base URL to the documented OpenAI-compatible endpoint while retaining compatibility with the previous root URL.
- Fixed the duplicated `Replicate` channel key in the `berry` theme channel constants.
- Serialized frontend theme builds in the Docker image build so the shared build directory is not corrupted by concurrent React builds.

### Fixed
- Refreshed the in-memory channel cache immediately after channel create, update, delete, and status changes so newly configured channels are available without waiting for the periodic sync loop.
- Invalidated cached per-group model lists after channel changes so `/v1/models` and request routing stay consistent after channel edits.

### Docs
- Documented Alibaba Bailian and Coding Plan channel setup and usage constraints in the README.
- Documented the channel cache refresh behavior in the Chinese and English FAQ sections.
