#!/bin/sh

if [ "$RUN_ENG_TRANSLATE" = "1" ]; then
  python ./i18n/translate.py --repository_path . --json_file_path ./i18n/en.json
else
  echo "RUN_ENG_TRANSLATE is not set to 1. We will not translate project to English (不会将项目翻译成英文)."
fi
