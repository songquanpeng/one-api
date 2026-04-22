# OneAPI Frontend i18n Status (DereineRing)

## Ist bereits mehrsprachig?

Ja, **teilweise**:

- `web/default` hat i18n mit `react-i18next`
- bereits vorhanden: `zh`, `en`
- jetzt erweitert: `de`, `sr` (initial als EN-basierte Fallback-Strings)

## Aktueller Stand in dieser Branch

Erweiterte Dateien:

- `web/default/src/i18n.js`
- `web/default/src/components/Header.js`
- `web/default/src/locales/de/translation.json`
- `web/default/src/locales/sr/translation.json`
- `web/default/src/locales/zh/translation.json`
- `scripts/i18n_audit.py`
- `docs/I18N_AUDIT.md`

Sprachen im Header-Dropdown:

- 中文 (`zh`)
- English (`en`)
- Deutsch (`de`)
- Srpski (`sr`)

## Was noch offen ist

- vollständige manuelle Übersetzung `de` und `sr` (ein großer Teil ist noch EN-Fallback)
- sprachliche Feinkorrektur der neu übersetzten Info-/Hinweistexte im Review
- optional: `web/berry` und `web/air` ebenfalls auf i18n umstellen

## Qualitätsprüfung

- Ein Audit-Skript prüft:
  - Key-Vollständigkeit je Sprache
  - Placeholder-Konsistenz (`{{name}}`, etc.)
  - Fortschrittsheuristik (`de/sr` identisch zu `en`)
- Aktueller Audit-Report: `docs/I18N_AUDIT.md`

## Warum dieser Ansatz?

So ist die Public-Branch sofort nutzbar und stabil (kein Build-Bruch), ohne private Daten preiszugeben.
