# Public Workspace Policy (DereineRing)

Dieses Repo ist für späteren Public Upload vorbereitet.

## Erlaubt (public)

- Quellcode für OneAPI i18n
- allgemeine Build- und Deploy-Dokumentation
- Platzhalter-Konfigurationen (`.env.example`, Templates)

## Verboten (nicht public)

- echte API Keys / Tokens
- Matrix Access Tokens
- lokale Benutzer-/Serverdaten
- private Infrastrukturpfade und persönliche Zugangsdaten

## Release-Checkliste vor Push/Fork

1. `git status` prüfen
2. `git diff` auf Secrets prüfen
3. Keine `.env` / `data/` / Logs commiten
4. Nur Template-Dateien veröffentlichen

## Trennung Public vs. Private

- **Public Entwicklung:** `/Volumes/M4Data/Coding/DereineRing/one-api`
- **Private Runtime/Secrets:** `/Volumes/M4Data/Coding/CommandStack-Private`
