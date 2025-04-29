#!/bin/bash

set -euo pipefail

if [ $# -lt 1 ]; then
  echo "❌ Usage: $0 input.ts [options]"
  echo "Options:"
  echo "  --hls             Generate regular HLS (playlist.m3u8 + segments/)"
  echo "  --sprite          Generate preview sprites (split every ~1 hour)"
  echo "  --previewmp4      Generate low-res preview.mp4 after sprites"
  echo "  --keep            Keep original .ts file after processing"
  echo "  --slow            Run ffmpeg in low-CPU mode (nice + low threads)"
  echo "  --all             Do everything (default)"
  exit 1
fi

INPUT="$1"
shift

# Flags
DO_HLS=false
DO_SPRITE=false
DO_PREVIEW_MP4=false
KEEP_INPUT=false
LOW_CPU_MODE=false

if [ $# -eq 0 ]; then
  DO_HLS=true
  DO_SPRITE=true
  DO_PREVIEW_MP4=true
else
  for arg in "$@"; do
    case $arg in
      --hls) DO_HLS=true ;;
      --sprite) DO_SPRITE=true ;;
      --previewmp4) DO_PREVIEW_MP4=true ;;
      --keep) KEEP_INPUT=true ;;
      --slow) LOW_CPU_MODE=true ;;
      --all) DO_HLS=true; DO_SPRITE=true; DO_PREVIEW_MP4=true ;;
      *) echo "❌ Unknown option: $arg"; exit 1 ;;
    esac
  done
fi

# Validate input
if [[ "$INPUT" != *.ts ]]; then
  echo "❌ Only .ts files are supported"
  exit 1
fi

BASENAME=$(basename "$INPUT" .ts)
OUTPUT_DIR="./$BASENAME"
LOCK_FILE="/tmp/build_media.lock"
LOG_FILE="$OUTPUT_DIR/build.log"
SPRITES_DIR="$OUTPUT_DIR/sprites"
JSON_PATH="$SPRITES_DIR/sprites.json"
VTT_PATH="$SPRITES_DIR/thumbnails.vtt"

THUMB_WIDTH=480
THUMB_HEIGHT=270
TILES_X=30
TILES_Y=24

# Блокировка выполнения
if [ -e "$LOCK_FILE" ]; then
  echo "❌ Another instance is already running (lock file exists: $LOCK_FILE)"
  exit 1
fi
trap 'rm -f "$LOCK_FILE"' EXIT
touch "$LOCK_FILE"

# Проверка существования выходной папки
if [ -d "$OUTPUT_DIR" ]; then
  echo "❌ Output directory $OUTPUT_DIR already exists. Aborting to prevent overwrite."
  exit 1
fi

mkdir -p "$OUTPUT_DIR"

# Setup ffmpeg prefix
if $LOW_CPU_MODE; then
  echo "⚡ Running in low-CPU mode (nice + limited threads)..."
  FFMPEG_PREFIX="nice -n 10 ffmpeg -threads 2 -nostdin"
else
  FFMPEG_PREFIX="ffmpeg -nostdin"
fi

# Utilities
current_start_time=0

log_start() {
  echo "▶️  Starting $1..." | tee -a "$LOG_FILE"
  date "+%Y-%m-%d %H:%M:%S" | tee -a "$LOG_FILE"
  current_start_time=$(date +%s)
}

log_done() {
  local now=$(date +%s)
  local elapsed=$((now - current_start_time))
  echo "✅ Finished $1 (⏱ ${elapsed}s)" | tee -a "$LOG_FILE"
  echo "" >> "$LOG_FILE"
}

log_fail() {
  local now=$(date +%s)
  local elapsed=$((now - current_start_time))
  echo "❌ Failed $1 (⏱ ${elapsed}s)" | tee -a "$LOG_FILE"
  echo "" >> "$LOG_FILE"
}

# Processing
if $DO_HLS; then
  log_start "HLS generation"
  mkdir -p "$OUTPUT_DIR/segments"

  if eval $FFMPEG_PREFIX -i "$INPUT" \
    -c:v copy \
    -c:a copy \
    -hls_time 5 \
    -hls_segment_type mpegts \
    -hls_flags independent_segments \
    -hls_segment_filename "$OUTPUT_DIR/segments/%d.ts" \
    -hls_base_url "segments/" \
    -hls_list_size 0 \
    -f hls "$OUTPUT_DIR/playlist.m3u8"; then
    log_done "HLS generation"
  else
    log_fail "HLS generation"
  fi
fi

if $DO_SPRITE; then
  log_start "Sprite and VTT generation"
  mkdir -p "$SPRITES_DIR"

  DURATION=$(ffprobe -v error -select_streams v:0 -show_entries format=duration -of csv=p=0 "$INPUT" | cut -d'.' -f1)
  if [[ -z "$DURATION" ]]; then
    echo "❌ Failed to detect duration"
    exit 1
  fi

  FRAMES=$((DURATION / 5))
  FRAMES_PER_SPRITE=720 # Спрайт ≈ 1 час

  echo "ℹ️  Video duration: $DURATION sec, frames: ~$FRAMES, frames per sprite: $FRAMES_PER_SPRITE" | tee -a "$LOG_FILE"

  echo "WEBVTT" > "$VTT_PATH"
  echo "" >> "$VTT_PATH"

  JSON="{ \"sprites\": ["

  PARTS=$(( (FRAMES + FRAMES_PER_SPRITE - 1) / FRAMES_PER_SPRITE ))

  for ((i=0; i<PARTS; i++)); do
    START=$((i * FRAMES_PER_SPRITE * 5))
    SPRITE_FILE="sprite_${i}.jpg"
    echo "🎯 Generating sprite part $i (starting from ${START}s)..." | tee -a "$LOG_FILE"

    if eval $FFMPEG_PREFIX -ss "$START" -i "$INPUT" \
      -vf fps=1/5,scale=${THUMB_WIDTH}:-1,tile=${TILES_X}x${TILES_Y} \
      -frames:v 1 \
      -q:v 2 \
      "$SPRITES_DIR/$SPRITE_FILE"; then
      JSON+=" {\"file\": \"$SPRITE_FILE\", \"start_sec\": $START },"

      frame=0
      for ((y=0; y<$TILES_Y; y++)); do
        for ((x=0; x<$TILES_X; x++)); do
          sec=$((START + frame * 5))
          start_time=$(printf "%02d:%02d:%02d.000" $((sec/3600)) $(( (sec%3600)/60 )) $((sec%60)))
          end_sec=$((sec+5))
          end_time=$(printf "%02d:%02d:%02d.000" $((end_sec/3600)) $(( (end_sec%3600)/60 )) $((end_sec%60)))
          offset_x=$((x * THUMB_WIDTH))
          offset_y=$((y * THUMB_HEIGHT))
          echo "$start_time --> $end_time" >> "$VTT_PATH"
          echo "$SPRITE_FILE#xywh=${offset_x},${offset_y},${THUMB_WIDTH},${THUMB_HEIGHT}" >> "$VTT_PATH"
          echo "" >> "$VTT_PATH"
          frame=$((frame+1))
        done
      done

    else
      log_fail "Sprite part $i"
    fi
  done

  JSON="${JSON%,}" # remove trailing comma
  JSON+=" ] }"
  echo "$JSON" > "$JSON_PATH"

  echo "✅ JSON index written to $JSON_PATH" | tee -a "$LOG_FILE"
  echo "✅ VTT index written to $VTT_PATH" | tee -a "$LOG_FILE"
  log_done "Sprite and VTT generation"
fi

if $DO_PREVIEW_MP4; then
  log_start "Preview MP4 generation"

  if eval $FFMPEG_PREFIX -i "$INPUT" \
    -vf "scale=640:360" \
    -an \
    -r 10 \
    -c:v libx264 \
    -preset veryfast \
    -crf 28 \
    -b:v 400k \
    "$OUTPUT_DIR/preview.mp4"; then
    log_done "Preview MP4 generation"
  else
    log_fail "Preview MP4 generation"
  fi
fi

# Remove input if not keeping
if ! $KEEP_INPUT; then
  echo "🗑️  Deleting original file: $INPUT" | tee -a "$LOG_FILE"
  rm -f "$INPUT"
else
  echo "📁 Keeping original file: $INPUT" | tee -a "$LOG_FILE"
fi