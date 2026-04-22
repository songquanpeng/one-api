# DerEineRing (EN)

## What this project is
DerEineRing is a public customization of OneAPI with a polished multilingual frontend.

It focuses on:
- UI localization: Chinese, English, German, Serbian
- cleaner public docs and translation workflow
- theme and UX polish in the web interface

## Who this is for
- teams running OneAPI as a local gateway
- users who want a multilingual admin UI
- contributors working on i18n and frontend quality

## Included languages
- `zh`
- `en`
- `de`
- `sr`

Locale files:
- `web/default/src/locales/zh/translation.json`
- `web/default/src/locales/en/translation.json`
- `web/default/src/locales/de/translation.json`
- `web/default/src/locales/sr/translation.json`

## Quick install (Docker)
```bash
git clone https://github.com/olivilo/one-api.git
cd one-api
docker compose up -d --build
```

Open:
- `http://localhost:3000`

## First login
On a fresh database, OneAPI usually creates:
- username: `root`
- password: `123456`

Change the password immediately after first login.

## How to use
1. Login to the web UI.
2. Add providers/channels in **Channels**.
3. Create an access token in **Tokens**.
4. Use OneAPI as an OpenAI-compatible endpoint:
   - Base URL: `http://localhost:3000/v1`
   - API key: your token from the UI

Example:
```bash
curl http://localhost:3000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ONEAPI_TOKEN" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role":"user","content":"Hello from DerEineRing"}]
  }'
```

## UI language and theme
- Use the language selector in the header to switch `zh/en/de/sr`.
- Use the theme selector in the header to switch available UI themes.

## Contributing
- Keep text changes in the locale JSON files.
- Keep terminology consistent across all four languages.
- Submit translation and UX changes via pull request.
