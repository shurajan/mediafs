#!/bin/bash

set -e

echo "📁 Создаю папку testdata/assets..."
mkdir -p testdata/assets

# Текстовые файлы
echo "📄 Создаю текстовые файлы..."
echo "hello world" > testdata/assets/hello.txt
echo "lorem ipsum dolor sit amet" > testdata/assets/lorem.txt
echo "this is a test file" > testdata/assets/test.txt

# Картинки
echo "🖼 Загружаю изображения..."

curl -fsSL -o testdata/assets/sample.jpg https://placehold.co/64x64.jpg \
  && echo "✅ sample.jpg готов" || echo "❌ Ошибка загрузки sample.jpg"

curl -fsSL -o testdata/assets/sample.png https://placehold.co/64x64.png \
  && echo "✅ sample.png готов" || echo "❌ Ошибка загрузки sample.png"

# Видео
echo "🎬 Генерация sample.mp4 (10 сек)..."
if command -v ffmpeg > /dev/null; then
  ffmpeg -loglevel error -f lavfi -i testsrc=duration=10:size=128x128:rate=1 \
    -c:v libx264 -t 10 -pix_fmt yuv420p testdata/assets/sample.mp4 \
    && echo "✅ sample.mp4 готов" || echo "❌ Ошибка создания sample.mp4"
else
  echo "⚠️ ffmpeg не установлен — sample.mp4 не будет создан"
fi

echo "✅ Все тестовые данные подготовлены в ./testdata/assets"