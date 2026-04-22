#!/usr/bin/env python3
import json
import re
import sys
import time
from pathlib import Path
from typing import Any

from deep_translator import GoogleTranslator

ROOT = Path(__file__).resolve().parents[1]
EN_PATH = ROOT / "web/default/src/locales/en/translation.json"

PLACEHOLDER_RE = re.compile(r"\{\{[^{}]+\}\}|<[^>]+>|`[^`]+`|https?://\S+")
TOKEN_FMT = "§§{:03d}§§"


def mask(text: str):
    holders = []

    def repl(m):
        idx = len(holders)
        holders.append(m.group(0))
        return TOKEN_FMT.format(idx)

    masked = PLACEHOLDER_RE.sub(repl, text)
    return masked, holders


def unmask(text: str, holders):
    out = text
    for i, h in enumerate(holders):
        out = out.replace(TOKEN_FMT.format(i), h)
    return out


def walk_strings(node: Any, path=()):
    if isinstance(node, dict):
        for k, v in node.items():
            yield from walk_strings(v, path + (k,))
    elif isinstance(node, list):
        for i, v in enumerate(node):
            yield from walk_strings(v, path + (str(i),))
    elif isinstance(node, str):
        yield path, node


def set_at_path(root: Any, path, value: str):
    cur = root
    for p in path[:-1]:
        cur = cur[int(p)] if isinstance(cur, list) else cur[p]
    last = path[-1]
    if isinstance(cur, list):
        cur[int(last)] = value
    else:
        cur[last] = value


def translate_text(translator: GoogleTranslator, text: str, cache: dict):
    if text.strip() == "":
        return text
    if text in cache:
        return cache[text]

    masked, holders = mask(text)
    # Keep technical literals unchanged (model IDs, URLs, version tags, code-like tokens).
    if (
        (" " not in masked and re.search(r"[/:_@]|\\d", masked))
        or masked.startswith("gpt")
        or masked.startswith("gemini")
        or masked.startswith("llama")
        or masked.startswith("mistral")
    ):
        cache[text] = text
        return text

    for attempt in range(4):
        try:
            t = translator.translate(masked)
            if not isinstance(t, str) or t.strip() == "":
                t = text
            t = unmask(t, holders)
            cache[text] = t
            return t
        except Exception:
            if attempt == 3:
                cache[text] = text
                return text
            time.sleep(0.8 * (attempt + 1))


def run(lang_code: str, out_path: Path):
    en = json.loads(EN_PATH.read_text(encoding="utf-8"))
    out = json.loads(json.dumps(en, ensure_ascii=False))
    translator = GoogleTranslator(source="en", target=lang_code)
    cache = {}

    items = list(walk_strings(en))
    total = len(items)
    for idx, (path, text) in enumerate(items, 1):
        translated = translate_text(translator, text, cache)
        set_at_path(out, path, translated)
        if idx % 80 == 0 or idx == total:
            print(f"[{lang_code}] {idx}/{total}")

    out_path.write_text(json.dumps(out, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"written: {out_path}")


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("usage: translate_locales.py <lang_code> <output_file>")
        sys.exit(1)
    run(sys.argv[1], Path(sys.argv[2]))
