package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// linux
	xclip  = "xclip"
	wlcopy = "wl-copy"
	xsell  = "xsel"
	// windows
	clip = "clip"
	// mac
	pbcopy = "pbcopy"
)

const (
	colorReset = "\033[0m"
	colorGreen = "\033[32m"
	colorWhite = "\033[97m"
	colorCyan  = "\033[36m"
)

const tmuxPreviousMode = "@go_previous_pane_mode"

var logger *log.Logger

func init() {
	logDir := "/tmp/limbo"
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		logger = log.New(os.Stderr, "LIMBO_DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
		logger.Printf("ERRO: Não foi possível criar diretório de log %s: %v", logDir, err)
		return
	}

	logPath := filepath.Join(logDir, "arquivo.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		logger = log.New(os.Stderr, "LIMBO_DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
		logger.Printf("ERRO: Não foi possível abrir arquivo de log %s: %v", logPath, err)
		return
	}

	logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Printf("=== INICIANDO SESSÃO DE DEBUG ===")
	logger.Printf("Sistema: %s %s", runtime.GOOS, runtime.GOARCH)
}

func main() {
	logger.Printf("main(): Iniciando aplicação")

	if err := run(); err != nil {
		logger.Printf("main(): ERRO - %v", err)
		_ = displayPopup("Error: " + err.Error())
		os.Exit(1)
	}

	logger.Printf("main(): Finalizado com sucesso")
}

func run() error {
	logger.Printf("run(): Iniciando execução principal")

	if err := ensureTmux(); err != nil {
		logger.Printf("run(): ERRO ensureTmux - %v", err)
		return err
	}
	logger.Printf("run(): Tmux verificado com sucesso")

	currentMode, err := getTmuxMode()
	if err != nil {
		logger.Printf("run(): ERRO getTmuxMode - %v", err)
		return err
	}
	logger.Printf("run(): Modo atual do tmux: '%s'", currentMode)

	defer func() {
		logger.Printf("run(): Definindo opção anterior: %s=%s", tmuxPreviousMode, currentMode)
		_ = setTmuxOption(tmuxPreviousMode, currentMode)
	}()

	previous, _ := getTmuxOption(tmuxPreviousMode)
	logger.Printf("run(): Modo anterior: '%s'", previous)

	justLeftCopyMode := strings.HasPrefix(previous, "copy-mode") &&
		!strings.HasPrefix(currentMode, "copy-mode")

	logger.Printf("run(): justLeftCopyMode = %v (anterior: '%s', atual: '%s')",
		justLeftCopyMode, previous, currentMode)

	if !justLeftCopyMode {
		logger.Printf("run(): Não saiu do modo de cópia recentemente, saindo")
		return nil
	}

	logger.Printf("run(): Detectada saída do modo de cópia, processando buffer...")

	text, err := getTmuxBuffer()
	if err != nil {
		logger.Printf("run(): ERRO getTmuxBuffer - %v", err)
		return err
	}
	logger.Printf("run(): Buffer obtido, tamanho: %d caracteres", len(text))

	clipboardTool, exists := hasClipboardTool()
	if !exists {
		logger.Printf("run(): ERRO - Nenhuma ferramenta de clipboard encontrada")
		return errors.New("no clipboard tool found")
	}
	logger.Printf("run(): Ferramenta de clipboard selecionada: %s", clipboardTool)

	if err := sendToClipboard(clipboardTool, text); err != nil {
		logger.Printf("run(): ERRO sendToClipboard - %v", err)
		return err
	}
	logger.Printf("run(): Texto enviado para clipboard com sucesso")

	showSuccess(text)
	logger.Printf("run(): Sucesso exibido ao usuário")

	return nil
}

func ensureTmux() error {
	logger.Printf("ensureTmux(): Verificando se tmux está disponível")
	_, err := exec.LookPath("tmux")
	if err != nil {
		logger.Printf("ensureTmux(): ERRO - tmux não encontrado")
		return errors.New("tmux not found")
	}
	logger.Printf("ensureTmux(): Tmux encontrado")
	return nil
}

func getTmuxMode() (string, error) {
	logger.Printf("getTmuxMode(): Executando comando tmux display-message")
	cmd := exec.Command("tmux", "display-message", "-p", "#{pane_mode}")
	out, err := cmd.Output()
	if err != nil {
		logger.Printf("getTmuxMode(): ERRO - %v", err)
		return "", err
	}
	result := strings.TrimSpace(string(out))
	logger.Printf("getTmuxMode(): Resultado: '%s'", result)
	return result, nil
}

func getTmuxOption(option string) (string, error) {
	logger.Printf("getTmuxOption(): Obtendo opção: %s", option)
	cmd := exec.Command("tmux", "show-options", "-gqv", option)
	out, err := cmd.Output()
	if err != nil {
		logger.Printf("getTmuxOption(): ERRO - %v", err)
		return "", nil
	}
	result := strings.TrimSpace(string(out))
	logger.Printf("getTmuxOption(): %s = '%s'", option, result)
	return result, nil
}

func setTmuxOption(option, value string) error {
	logger.Printf("setTmuxOption(): Definindo %s = '%s'", option, value)
	cmd := exec.Command("tmux", "set-option", "-g", option, value)
	err := cmd.Run()
	if err != nil {
		logger.Printf("setTmuxOption(): ERRO - %v", err)
	} else {
		logger.Printf("setTmuxOption(): Sucesso")
	}
	return err
}

func getTmuxBuffer() (string, error) {
	logger.Printf("getTmuxBuffer(): Obtendo buffer do tmux")
	cmd := exec.Command("tmux", "show-buffer")
	out, err := cmd.Output()
	if err != nil {
		logger.Printf("getTmuxBuffer(): ERRO - %v", err)
		return "", errors.New("there is nothing in the buffer")
	}
	text := string(out)
	if text == "" {
		logger.Printf("getTmuxBuffer(): ERRO - Buffer vazio")
		return "", errors.New("no text selected")
	}

	logger.Printf("getTmuxBuffer(): Buffer obtido, primeiros 100 chars: '%s'",
		truncateText(text, 100))
	return text, nil
}

func hasClipboardTool() (string, bool) {
	logger.Printf("hasClipboardTool(): Procurando ferramentas de clipboard")
	tools := getPlatformTools()
	logger.Printf("hasClipboardTool(): Ferramentas disponíveis para %s: %v", runtime.GOOS, tools)

	for _, tool := range tools {
		if hasBinary(tool) {
			logger.Printf("hasClipboardTool(): Encontrada: %s", tool)
			return tool, true
		}
		logger.Printf("hasClipboardTool(): %s não encontrada", tool)
	}

	logger.Printf("hasClipboardTool(): Nenhuma ferramenta encontrada")
	return "", false
}

func getPlatformTools() []string {
	var tools []string
	switch runtime.GOOS {
	case "linux":
		tools = []string{xclip, wlcopy, xsell}
	case "windows":
		tools = []string{clip}
	case "darwin":
		tools = []string{pbcopy}
	default:
		tools = nil
	}
	logger.Printf("getPlatformTools(): SO=%s, tools=%v", runtime.GOOS, tools)
	return tools
}

func sendToClipboard(tool, text string) error {
	logger.Printf("sendToClipboard(): Enviando para %s, texto tamanho: %d", tool, len(text))
	args := getToolArgs(tool)
	logger.Printf("sendToClipboard(): Argumentos: %v", args)
	return pipTo(tool, text, args...)
}

func getToolArgs(tool string) []string {
	var zero []string
	var args []string

	switch tool {
	case xclip:
		args = []string{"-selection", "clipboard"}
	case wlcopy:
		args = zero
	case xsell:
		args = []string{"--clipboard", "--input"}
	default:
		args = zero
	}

	logger.Printf("getToolArgs(): %s -> %v", tool, args)
	return args
}

func pipTo(tool, text string, args ...string) error {
	logger.Printf("pipTo(): Executando %s com %d argumentos", tool, len(args))

	cmd := exec.Command(tool, args...)
	logger.Printf("pipTo(): Comando: %s %s", tool, strings.Join(args, " "))

	input, err := cmd.StdinPipe()
	if err != nil {
		logger.Printf("pipTo(): ERRO ao criar stdin pipe - %v", err)
		return err
	}

	if err := cmd.Start(); err != nil {
		logger.Printf("pipTo(): ERRO ao iniciar comando - %v", err)
		return err
	}
	logger.Printf("pipTo(): Comando iniciado com PID %d", cmd.Process.Pid)

	start := time.Now()
	_, _ = io.WriteString(input, text)
	_ = input.Close()
	err = cmd.Wait()
	duration := time.Since(start)

	if err != nil {
		logger.Printf("pipTo(): ERRO ao aguardar comando - %v (duração: %v)", err, duration)
	} else {
		logger.Printf("pipTo(): Comando finalizado com sucesso (duração: %v)", duration)
	}

	return err
}

func hasBinary(name string) bool {
	logger.Printf("hasBinary(): Verificando se %s existe no PATH", name)
	_, err := exec.LookPath(name)
	exists := err == nil
	logger.Printf("hasBinary(): %s -> %v", name, exists)
	return exists
}

func displayPopup(text string) error {
	logger.Printf("displayPopup(): Exibindo popup, texto tamanho: %d", len(text))
	colored := fmt.Sprintf(
		"%s ✔ Copied text!%s\n\n"+
			"%s%s%s\n\n"+
			"%s Press ENTER to close...%s",
		colorGreen, colorReset,
		colorWhite, text, colorReset,
		colorCyan, colorReset,
	)

	err := exec.Command(
		"tmux", "display-popup", "-E",
		fmt.Sprintf("printf \"%s\"; read _", escapeForShell(colored)),
	).Run()

	if err != nil {
		logger.Printf("displayPopup(): ERRO - %v", err)
	} else {
		logger.Printf("displayPopup(): Popup exibido com sucesso")
	}

	return err
}

func escapeForShell(s string) string {
	result := strings.ReplaceAll(s, "\"", "\\\"")
	logger.Printf("escapeForShell(): Entrada: %d chars, Saída: %d chars", len(s), len(result))
	return result
}

func showSuccess(text string) {
	logger.Printf("showSuccess(): Preparando mensagem de sucesso")
	preview := text
	if len(text) > 350 {
		preview = text[:350] + "..."
		logger.Printf("showSuccess(): Texto truncado de %d para 350 caracteres", len(text))
	}

	msg := fmt.Sprintf(
		"✔ Texto copiado!\n\n%s",
		preview,
	)

	logger.Printf("showSuccess(): Mensagem preparada, tamanho: %d caracteres", len(msg))
	_ = displayPopup(msg)
}

func truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}
