#!/usr/bin/env python3
import json
import re
from pathlib import Path

BASE = Path(__file__).resolve().parents[1] / 'web' / 'default' / 'src' / 'locales'
LANGS = ['zh', 'en', 'de', 'sr']
PH_RE = re.compile(r'\{\{\s*([a-zA-Z0-9_]+)\s*\}\}')


def flatten(obj, prefix=''):
    out = {}
    if isinstance(obj, dict):
        for k, v in obj.items():
            p = f"{prefix}.{k}" if prefix else k
            if isinstance(v, dict):
                out.update(flatten(v, p))
            else:
                out[p] = v
    return out


def placeholders(text):
    if not isinstance(text, str):
        return set()
    return set(PH_RE.findall(text))


def load_lang(lang):
    p = BASE / lang / 'translation.json'
    return flatten(json.loads(p.read_text(encoding='utf-8')))


def main():
    data = {l: load_lang(l) for l in LANGS}
    base_keys = set(data['en'])

    print('# i18n Audit Report')
    print('')
    print(f'- Languages: {", ".join(LANGS)}')
    print(f'- Base locale: en ({len(base_keys)} keys)')
    print('')

    print('## Key Coverage')
    for lang in LANGS:
        keys = set(data[lang])
        missing = sorted(base_keys - keys)
        extra = sorted(keys - base_keys)
        print(f'- {lang}: keys={len(keys)}, missing={len(missing)}, extra={len(extra)}')

    print('')
    print('## Placeholder Consistency (vs en)')
    for lang in LANGS:
        if lang == 'en':
            continue
        mismatch = []
        for k in sorted(base_keys):
            en_ph = placeholders(data['en'].get(k, ''))
            lg_ph = placeholders(data[lang].get(k, ''))
            if en_ph != lg_ph:
                mismatch.append((k, en_ph, lg_ph))
        print(f'- {lang}: mismatches={len(mismatch)}')
        for k, en_ph, lg_ph in mismatch[:10]:
            print(f'  - {k}: en={sorted(en_ph)} {lang}={sorted(lg_ph)}')

    print('')
    print('## Translation Progress Heuristic (identical to en)')
    for lang in ['de', 'sr']:
        identical = 0
        for k in base_keys:
            if isinstance(data['en'].get(k), str) and data[lang].get(k) == data['en'].get(k):
                identical += 1
        ratio = identical / len(base_keys) * 100
        print(f'- {lang}: identical_to_en={identical}/{len(base_keys)} ({ratio:.1f}%)')


if __name__ == '__main__':
    main()
