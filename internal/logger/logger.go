// Package logger предоставляет глобальные логгеры для разных уровней логирования.
// Используется во всех частях микросервиса без подключения сторонних библиотек.
// в .env DEBUG=true — для разработки. DEBUG=false — для продакшена.
package logger

import (
	"log"
	"os"

	"order_service/config"
)

var (
	// InfoLogger используется для логирования общей информации
	// о ходе работы приложения (запуск, действия пользователя и т.п.).
	InfoLogger *log.Logger

	// ErrorLogger выводит сообщения об ошибках, которые требуют внимания.
	ErrorLogger *log.Logger

	// DebugLogger предназначен для отладки — можно отключить в продакшене.
	DebugLogger *log.Logger
)

// InitLogger инициализирует глобальные логгеры.
// Достаточно вызвать один раз при запуске приложения (например, в main).
func InitLogger(cfg *config.Config) {
	debug := cfg.Serv.Debug

	InfoLogger = log.New(os.Stdout, "[INFO]  ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)

	if debug {
		DebugLogger = log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile)
	} else {
		// Открываем /dev/null как io.Writer
		nullFile, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		if err != nil {
			// На всякий случай, чтобы не упасть, отправим в stderr
			DebugLogger = log.New(os.Stderr, "[DEBUG-OFF-FAILSAFE] ", log.Ldate|log.Ltime|log.Lshortfile)

			return
		}

		DebugLogger = log.New(nullFile, "", 0)
	}
}

// Info логирует информационное сообщение.
func Info(message string) {
	InfoLogger.Println(message)
}

// Warn логирует предупреждение.
func Warn(message string) {
	InfoLogger.Println("[WARNING] " + message)
}

// Error логирует сообщение об ошибке.
func Error(message string) {
	ErrorLogger.Println(message)
}

// Debug logs a debug-level message using DebugLogger.
func Debug(message string) {
	DebugLogger.Println(message)
}
