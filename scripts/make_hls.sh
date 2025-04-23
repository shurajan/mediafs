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

echo "📦 Converting $INPUT to HLS (no re-encoding) → $OUTPUT_DIR"

mkdir -p "$OUTPUT_DIR"

# 1️⃣ Попытка сохранить видео и аудио как есть
echo "🎧 Trying with original audio..."
if ffmpeg -i "$INPUT" \
  -c:v copy \
  -c:a aac -b:a 128k -ac 2 -ar 48000 -strict -2 \
  -hls_time 10 \
  -hls_segment_type mpegts \
  -hls_segment_filename "$OUTPUT_DIR/%d.ts" \
  -hls_list_size 0 \
  -f hls "$OUTPUT_DIR/playlist.m3u8"; then

  echo "✅ HLS (video copy, audio AAC) created successfully in $OUTPUT_DIR"

else
  echo "⚠️  Audio processing failed. Retrying without audio..."

  # 2️⃣ Fallback: видео сохраняем, аудио вырезаем
  if ffmpeg -i "$INPUT" \
    -c:v copy \
    -an \
    -hls_time 10 \
    -hls_segment_type mpegts \
    -hls_segment_filename "$OUTPUT_DIR/%d.ts" \
    -hls_list_size 0 \
    -f hls "$OUTPUT_DIR/playlist.m3u8"; then

    echo "✅ HLS created WITHOUT audio in $OUTPUT_DIR"
  else
    echo "❌ Failed to generate HLS with or without audio."
    exit 1
  fi
fi

rm -f "$INPUT"
echo "🗑️  Deleted original file: $INPUT"