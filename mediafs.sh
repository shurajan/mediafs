#!/bin/bash

# Путь к папке с проектом и исполняемым файлом
PROJECT_DIR="$HOME/projects/mediafs"
BUILD_OUTPUT="$PROJECT_DIR/mediafs"
SERVICE_DIR="$HOME/.mediafs/bin"
SERVICE_NAME="mediafs.service"
SERVICE_FILE="/etc/systemd/system/$SERVICE_NAME"

# Проверка и создание папки для исполняемого файла
if [ ! -d "$SERVICE_DIR" ]; then
  echo "Создание папки для бинарника: $SERVICE_DIR"
  mkdir -p "$SERVICE_DIR" || { echo "Не удалось создать папку $SERVICE_DIR"; exit 1; }
fi

echo "Переход в папку проекта..."
cd "$PROJECT_DIR" || { echo "Не удалось перейти в $PROJECT_DIR"; exit 1; }

echo "Обновление репозитория..."
git pull || { echo "Не удалось выполнить git pull"; exit 1; }

echo "Сборка проекта..."
go build -o "$BUILD_OUTPUT" ./cmd/mediafs || { echo "Ошибка при сборке проекта"; exit 1; }

# Проверка существования сервиса
if [ -f "$SERVICE_FILE" ]; then
  echo "Сервис $SERVICE_NAME найден. Остановка службы..."
  sudo systemctl stop "$SERVICE_NAME" || { echo "Не удалось остановить службу $SERVICE_NAME"; exit 1; }
else
  echo "Сервис $SERVICE_NAME не найден. Будет выполнена его установка."
fi

echo "Копирование бинарника MediaFS в $SERVICE_DIR..."
sudo cp "$BUILD_OUTPUT" "$SERVICE_DIR" || { echo "Ошибка при копировании файла"; exit 1; }

# Создание systemd unit файла, если он не существует
if [ ! -f "$SERVICE_FILE" ]; then
  echo "Создание файла службы $SERVICE_NAME..."
  sudo bash -c "cat > $SERVICE_FILE" <<EOF
[Unit]
Description=MediaFS Server
After=network.target

[Service]
ExecStart=$SERVICE_DIR/mediafs
Restart=always
User=$USER
WorkingDirectory=$SERVICE_DIR
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
EOF

  sudo systemctl daemon-reload || { echo "Ошибка при перезагрузке systemd"; exit 1; }
  sudo systemctl enable "$SERVICE_NAME" || { echo "Не удалось включить службу $SERVICE_NAME"; exit 1; }
fi

echo "Запуск службы $SERVICE_NAME..."
sudo systemctl start "$SERVICE_NAME" || { echo "Не удалось запустить службу $SERVICE_NAME"; exit 1; }

echo "MediaFS успешно обновлён и перезапущен."