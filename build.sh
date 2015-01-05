#!/bin/sh

OUT="./built/"
if [ ! -d "$OUT" ]; then
    mkdir "$OUT"
else
    rm -rf "$OUT"
    mkdir "$OUT"
fi

VERSION="$1"
if [ $# -eq 0 ]; then
    VERSION="dev"
fi

cd "$OUT"
gox -output "chameleon_${VERSION}_{{.OS}}_{{.Arch}}" github.com/NickPresta/chameleon

for file in chameleon_*; do
    if [[ "$file" == *.exe ]]; then
        name="chameleon.exe"
    else
        name="chameleon"
    fi
    mv "$file" $name
    zip "$file.zip" $name
    rm $name
done
