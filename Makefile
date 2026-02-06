# Версия приложения
VERSION=1.0.0

# Получение даты сборки
BUILD_DATE=$(shell date +"%Y-%m-%dT%H%M")

# Директории
BIN_DIR=./bin
CLIENT_BIN_DIR=$(BIN_DIR)/client
SERVER_BIN_DIR=$(BIN_DIR)/server

# Имена исполняемых файлов
CLIENT_EXE=gophkeeper-cli
SERVER_EXE=gophkeeper-server

# Пути к исполняемым файлам
ifeq ($(OS),Windows_NT)
    CLIENT_TARGET=$(CLIENT_BIN_DIR)/$(CLIENT_EXE).exe
    SERVER_TARGET=$(SERVER_BIN_DIR)/$(SERVER_EXE).exe
else
    CLIENT_TARGET=$(CLIENT_BIN_DIR)/$(CLIENT_EXE)
    SERVER_TARGET=$(SERVER_BIN_DIR)/$(SERVER_EXE)
endif

# Флаги сборки для клиента
CLIENT_LDFLAGS=-X 'gophkeeper/internal/services.Version=$(VERSION)' -X 'gophkeeper/internal/services.BuildDate=$(BUILD_DATE)'

# Цели по умолчанию
.PHONY: all client server clean help

all: client server

# Создание директорий
$(CLIENT_BIN_DIR) $(SERVER_BIN_DIR):
	@mkdir -p $@

# Сборка клиента
client: $(CLIENT_TARGET)

$(CLIENT_TARGET): $(CLIENT_BIN_DIR)
	@echo "Building GophKeeper Client..."
	go mod tidy
	go build -ldflags "$(CLIENT_LDFLAGS)" -o $(CLIENT_TARGET) ./cmd/client
	@echo "GophKeeper Client built successfully! ($(CLIENT_TARGET))"

# Сборка сервера
server: $(SERVER_TARGET)

$(SERVER_TARGET): $(SERVER_BIN_DIR)
	@echo "Building GophKeeper Server..."
	go mod tidy
	go build -o $(SERVER_TARGET) ./cmd/server/serv
	@echo "GophKeeper Server built successfully! ($(SERVER_TARGET))"

# Очистка
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)/*
	@echo "Clean completed!"

# Справка
help:
	@echo "Доступные команды:"
	@echo "  make all     - собрать клиент и сервер (цель по умолчанию)"
	@echo "  make client  - собрать только клиент"
	@echo "  make server  - собрать только сервер"
	@echo "  make clean   - удалить скомпилированные файлы"
	@echo "  make help    - показать эту справку"
	@echo ""
	@echo "Параметры:"
	@echo "  VERSION=$(VERSION)"
	@echo "  BUILD_DATE=$(BUILD_DATE)"