package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

var Logger *log.Logger

func init() {
	logDir, err := os.UserConfigDir()
	logDir = fmt.Sprintf("%s/%s", logDir, "tmux-copy-plugin")
	if err != nil {
		Logger = log.New(os.Stderr, "LIMBO_DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
		Logger.Printf("ERRO: Não foi possível encontrar diretório do usuário %s: %v", logDir, err)
		return
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		Logger = log.New(os.Stderr, "LIMBO_DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
		Logger.Printf("ERRO: Não foi possível criar diretório de log %s: %v", logDir, err)
		return
	}

	logPath := filepath.Join(logDir, "tmux-copy-plugin.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		Logger = log.New(os.Stderr, "LIMBO_DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
		Logger.Printf("ERRO: Não foi possível abrir arquivo de log %s: %v", logPath, err)
		return
	}

	Logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
	Logger.Printf("=== INICIANDO SESSÃO DE DEBUG ===")
	Logger.Printf("Sistema: %s %s", runtime.GOOS, runtime.GOARCH)
}
