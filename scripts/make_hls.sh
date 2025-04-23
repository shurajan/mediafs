#!/bin/bash

set -euo pipefail

if [ $# -lt 1 ]; then
  echo "❌ Usage: $0 input.ts"
  exit 1
fi

INPUT="$1"

# Проверка расширения
if [[ "$INPUT" != *.ts ]]; then
  echo "❌ Only .ts files are supported"
  exit 1
fi

# Имя директории = имя без расширения
BASENAME=$(basename "$INPUT" .ts)
OUTPUT_DIR="./$BASENAME"

echo "📦 Converting $INPUT to HLS → $OUTPUT_DIR"

mkdir -p "$OUTPUT_DIR"

# Генерация HLS с выровненными ключевыми кадрами каждые 10 секунд
ffmpeg -i "$INPUT" \
  -c:v libx264 -preset veryfast -crf 23 \
  -c:a aac -b:a 128k \
  -force_key_frames "expr:gte(t,n_forced*10)" \
  -hls_time 10 \
  -hls_segment_type mpegts \
  -hls_segment_filename "$OUTPUT_DIR/%d.ts" \
  -hls_list_size 0 \
  -f hls "$OUTPUT_DIR/playlist.m3u8"

echo "✅ HLS created successfully in $OUTPUT_DIR"

rm -f "$INPUT"
echo "🗑️  Deleted original file: $INPUT"