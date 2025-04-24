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

BASENAME=$(basename "$INPUT" .ts)
OUTPUT_DIR="./$BASENAME"

echo "📦 Converting $INPUT to HLS (no re-encoding) → $OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

PLAYLIST_PATH="$OUTPUT_DIR/playlist.m3u8"
SEGMENT_PATTERN="$OUTPUT_DIR/%d.ts"

# 1️⃣ Преобразование с сохранением видео, перекодированием аудио
echo "🎧 Trying with original video and AAC audio..."
if ffmpeg -i "$INPUT" \
  -c:v copy \
  -c:a aac -b:a 128k -ac 2 -ar 48000 -strict -2 \
  -hls_time 10 \
  -hls_segment_type mpegts \
  -hls_segment_filename "$SEGMENT_PATTERN" \
  -hls_list_size 0 \
  -f hls "$PLAYLIST_PATH"; then

  echo "✅ HLS (video copy, audio AAC) created successfully"

else
  echo "⚠️  Audio processing failed. Retrying without audio..."
  if ffmpeg -i "$INPUT" \
    -c:v copy \
    -an \
    -hls_time 10 \
    -hls_segment_type mpegts \
    -hls_segment_filename "$SEGMENT_PATTERN" \
    -hls_list_size 0 \
    -f hls "$PLAYLIST_PATH"; then

    echo "✅ HLS created WITHOUT audio"
  else
    echo "❌ Failed to generate HLS with or without audio."
    exit 1
  fi
fi

# 2️⃣ Извлечение разрешения
WIDTH=$(ffprobe -v error -select_streams v:0 -show_entries stream=width -of csv=p=0 "$INPUT" | head -n 1 | tr -d '\r\n')
HEIGHT=$(ffprobe -v error -select_streams v:0 -show_entries stream=height -of csv=p=0 "$INPUT" | head -n 1 | tr -d '\r\n')

if [[ -z "$WIDTH" || -z "$HEIGHT" ]]; then
  echo "❌ Failed to extract resolution from $INPUT"
  exit 1
fi

# 3️⃣ Создание master.m3u8 с RESOLUTION
MASTER_PATH="$OUTPUT_DIR/master.m3u8"
BANDWIDTH=5000000

cat <<EOF > "$MASTER_PATH"
#EXTM3U
#EXT-X-STREAM-INF:BANDWIDTH=$BANDWIDTH,RESOLUTION=${WIDTH}x${HEIGHT}
playlist.m3u8
EOF

# 4️⃣ Удаление исходного файла
rm -f "$INPUT"
echo "🗑️  Deleted original file: $INPUT"