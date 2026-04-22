# DerEineRing (DE)

## Was dieses Projekt ist
DerEineRing ist eine öffentliche OneAPI-Anpassung mit mehrsprachigem, überarbeitetem Frontend.

Der Fokus liegt auf:
- UI-Lokalisierung: Chinesisch, Englisch, Deutsch, Serbisch
- klarer öffentlicher Doku und Übersetzungs-Workflow
- visueller und funktionaler Frontend-Politur

## Für wen es gedacht ist
- Teams, die OneAPI lokal als Gateway betreiben
- Nutzer, die eine mehrsprachige Admin-Oberfläche wollen
- Contributor für i18n- und Frontend-Qualität

## Enthaltene Sprachen
- `zh`
- `en`
- `de`
- `sr`

Locale-Dateien:
- `web/default/src/locales/zh/translation.json`
- `web/default/src/locales/en/translation.json`
- `web/default/src/locales/de/translation.json`
- `web/default/src/locales/sr/translation.json`

## Schnellstart (Docker)
```bash
git clone https://github.com/olivilo/one-api.git
cd one-api
docker compose up -d --build
```

Aufruf:
- `http://localhost:3000`

## Erster Login
Bei einer frischen Datenbank legt OneAPI in der Regel an:
- Benutzername: `root`
- Passwort: `123456`

Passwort direkt nach dem ersten Login ändern.

## Nutzung
1. Im Web-UI anmelden.
2. Provider/Kanäle unter **Channels** anlegen.
3. Zugriffstoken unter **Tokens** erstellen.
4. OneAPI als OpenAI-kompatiblen Endpoint nutzen:
   - Base URL: `http://localhost:3000/v1`
   - API-Key: dein Token aus dem UI

Beispiel:
```bash
curl http://localhost:3000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ONEAPI_TOKEN" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role":"user","content":"Hallo von DerEineRing"}]
  }'
```

## Sprache und Theme im UI
- Über den Sprachumschalter im Header zwischen `zh/en/de/sr` wechseln.
- Über den Theme-Umschalter im Header zwischen den verfügbaren Designs wechseln.

## Mitwirken
- Textänderungen in den jeweiligen Locale-JSON-Dateien pflegen.
- Terminologie in allen vier Sprachen konsistent halten.
- Übersetzungs- und UX-Änderungen per Pull Request einreichen.
