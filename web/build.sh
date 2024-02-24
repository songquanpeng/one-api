#!/bin/sh

version=$(cat VERSION)
pwd

while IFS= read -r theme; do
    echo "Building theme: $theme"
    rm -r build/$theme
    cd "$theme"
    npm install
    DISABLE_ESLINT_PLUGIN='true' REACT_APP_VERSION=$version npm run build
    cd ..
done < THEMES
