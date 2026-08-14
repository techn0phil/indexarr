#!/bin/sh

set -eu

# Liste des extensions de fichiers connues
FILE_EXTENSIONS="mkv db txt sh png jpg jpeg mp4 nfo xml mpls m2ts clpi bdjo bdmv jar crt crl iso sig bin dat cci cer tbl lst inf cmap otf pcm upt xig xrl xrt prop hcf pdf zip cfg"

is_file() {
    name="$1"

    case "$name" in
        *.*)
            ext="${name##*.}"
            ext=$(printf '%s' "$ext" | tr '[:upper:]' '[:lower:]')

            for known in $FILE_EXTENSIONS; do
                [ "$ext" = "$known" ] && return 0
            done
            ;;
    esac

    return 1
}

# Lecture depuis un fichier ou stdin
if [ $# -gt 1 ]; then
    echo "Usage: $0 [ls-output.txt]" >&2
    exit 1
fi

if [ $# -eq 1 ]; then
    exec < "$1"
fi

current_dir="."

while IFS= read -r line || [ -n "$line" ]; do

    [ -z "$line" ] && continue

    case "$line" in

        *:)
            current_dir="${line%:}"
            mkdir -p "$current_dir"
            ;;

        *)
            path="$current_dir/$line"

            if is_file "$line"; then
                mkdir -p "$(dirname "$path")"
                touch "$path"
            else
                mkdir -p "$path"
            fi
            ;;
    esac

done
