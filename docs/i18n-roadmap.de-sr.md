# i18n Roadmap (DE/EN/SR/ZH)

## Ist-Stand
- OneAPI Frontend hat in `web/default/src/locales` aktuell nur:
  - `en/translation.json`
  - `zh/translation.json`
- Sprachumschalter ist in `web/default/src/components/Header.js` auf `zh` und `en` begrenzt.
- Andere Themes (`web/air`, `web/berry`) nutzen derzeit kein vollständiges i18n-Setup.

## Ziel
- Saubere Mehrsprachigkeit für:
  - Deutsch (`de`)
  - Englisch (`en`)
  - Serbisch (`sr`)
  - Chinesisch (`zh`)
- Konsistente Schlüsselstruktur, keine gemischten Hardcoded-Texte.

## Umsetzungsschritte
1. [x] Neue Locale-Dateien hinzufügen (`de`, `sr`) auf Basis des finalen Key-Sets.
2. [x] `web/default/src/i18n.js` um `de` und `sr` erweitern, Fallback auf `en` setzen.
3. [x] Sprachauswahl in `Header.js` um `Deutsch` und `Srpski` erweitern.
4. Diff-Check aller Keys zwischen `zh`, `en`, `de`, `sr` automatisieren.
5. UI-Screening pro Hauptseite (Login, Dashboard, Channel, Settings, Token, Logs).
6. Terminologie-Review (Agent, Channel, Token, Fallback, Quota, Billing).

## Status-Update (2026-04-20)
- Technische Mehrsprachigkeit ist aktiviert (ZH/EN/DE/SR auswählbar).
- `de` und `sr` sind aktuell noch auf EN-Basis und müssen fachlich sauber nachgezogen werden.

## Qualitätssicherung
- Key-Completeness-Check pro Sprache (100% Abdeckung).
- Visuelle Checks für lange Texte (DE/SR) auf Mobile/Desktop.
- Einheitliche Terminologie-Liste als Referenz.

## Aufwand (realistisch)
- Technische Integration: ~0.5 Tag
- Vollständige fachlich saubere Übersetzung + QA: ~1.5 bis 3 Tage
- Gesamt: ~2 bis 3.5 Tage für ein releasefähiges Ergebnis.

## Lizenz
- Repository ist MIT-lizenziert (Fork/Branch/public Beitrag ist zulässig).
