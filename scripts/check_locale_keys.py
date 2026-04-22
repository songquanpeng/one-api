#!/usr/bin/env python3
import json
from pathlib import Path

base = json.loads(Path('web/default/src/locales/en/translation.json').read_text())
langs = ['zh','de','sr']

def keys(obj,prefix=''):
    out=set()
    if isinstance(obj,dict):
        for k,v in obj.items():
            path=f'{prefix}.{k}' if prefix else k
            out.add(path)
            out |= keys(v,path)
    elif isinstance(obj,list):
        for i,v in enumerate(obj):
            path=f'{prefix}[{i}]'
            out.add(path)
            out |= keys(v,path)
    return out

base_keys=keys(base)
ok=True
for lang in langs:
    d=json.loads(Path(f'web/default/src/locales/{lang}/translation.json').read_text())
    k=keys(d)
    miss=sorted(base_keys-k)
    extra=sorted(k-base_keys)
    if miss or extra:
        ok=False
        print(f'[{lang}] missing={len(miss)} extra={len(extra)}')
        for x in miss[:10]: print('  MISSING',x)
        for x in extra[:10]: print('  EXTRA  ',x)
    else:
        print(f'[{lang}] keys OK ({len(k)})')

raise SystemExit(0 if ok else 1)
