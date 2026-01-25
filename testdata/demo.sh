#!/bin/bash
# Demo script - runs each sample through viewscreen

cd "$(dirname "$0")/.."

for f in testdata/*.jsonl; do
    echo -e "\n\033[1;36m━━━ $(basename "$f") ━━━\033[0m\n"
    cat "$f" | ./viewscreen -v
    echo
    read -p "Press Enter for next..."
done

echo -e "\n\033[1;32mDemo complete!\033[0m"
