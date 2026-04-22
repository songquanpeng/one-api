# DerEineRing (SR)

## Šta je ovaj projekat
DerEineRing je javna prilagođena verzija OneAPI-ja sa doteranim višejezičnim frontendom.

Fokus:
- lokalizovan UI: kineski, engleski, nemački, srpski
- jasna javna dokumentacija i prevodilački workflow
- unapređen izgled i UX interfejsa

## Kome je namenjen
- timovima koji lokalno koriste OneAPI kao gateway
- korisnicima kojima treba višejezični admin UI
- contributorima za i18n i frontend kvalitet

## Podržani jezici
- `zh`
- `en`
- `de`
- `sr`

Locale fajlovi:
- `web/default/src/locales/zh/translation.json`
- `web/default/src/locales/en/translation.json`
- `web/default/src/locales/de/translation.json`
- `web/default/src/locales/sr/translation.json`

## Brza instalacija (Docker)
```bash
git clone https://github.com/olivilo/one-api.git
cd one-api
docker compose up -d --build
```

Otvori:
- `http://localhost:3000`

## Prvo prijavljivanje
Na čistoj bazi OneAPI obično kreira:
- korisnik: `root`
- lozinka: `123456`

Promeni lozinku odmah nakon prvog logovanja.

## Kako se koristi
1. Uloguj se u web UI.
2. Dodaj provajdere/kanale u **Channels**.
3. Kreiraj pristupni token u **Tokens**.
4. Koristi OneAPI kao OpenAI-kompatibilni endpoint:
   - Base URL: `http://localhost:3000/v1`
   - API ključ: token iz UI-ja

Primer:
```bash
curl http://localhost:3000/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ONEAPI_TOKEN" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role":"user","content":"Pozdrav iz DerEineRing-a"}]
  }'
```

## Jezik i tema u UI-ju
- U headeru promeni jezik na `zh/en/de/sr`.
- U headeru promeni temu preko theme selektora.

## Doprinos projektu
- Tekstualne izmene radi kroz locale JSON fajlove.
- Održavaj doslednu terminologiju u sva četiri jezika.
- Šalji prevodilačke i UX izmene kroz pull request.
