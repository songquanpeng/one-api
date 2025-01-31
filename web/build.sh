#!/bin/sh

version=$(cat VERSION)
pwd

while IFS= read -r theme; do
    echo "Building theme: $theme"
    rm -r build/$theme
    cd "$theme"
    npm install
    jq ".homepage=\"${REACT_APP_BASE_URL}\"" package.json > tmp.json && mv tmp.json package.json ;
    DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$version npm run build
    cd ..
done < THEMES
