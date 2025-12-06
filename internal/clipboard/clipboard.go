package clipboard

import (
	"io"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/Marlliton/tmux-copy-plugin/internal/logger"
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

func HasTool() (string, bool) {
	logger.Logger.Printf("hasClipboardTool(): Procurando ferramentas de clipboard")
	tools := getPlatformTools()
	logger.Logger.Printf("hasClipboardTool(): Ferramentas disponíveis para %s: %v", runtime.GOOS, tools)

	for _, tool := range tools {
		if hasBinary(tool) {
			logger.Logger.Printf("hasClipboardTool(): Encontrada: %s", tool)
			return tool, true
		}
		logger.Logger.Printf("hasClipboardTool(): %s não encontrada", tool)
	}

	logger.Logger.Printf("hasClipboardTool(): Nenhuma ferramenta encontrada")
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
	logger.Logger.Printf("getPlatformTools(): SO=%s, tools=%v", runtime.GOOS, tools)
	return tools
}

func Send(tool, text string) error {
	logger.Logger.Printf("sendToClipboard(): Enviando para %s, texto tamanho: %d", tool, len(text))
	args := getToolArgs(tool)
	logger.Logger.Printf("sendToClipboard(): Argumentos: %v", args)
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
	case clip:
		args = zero
	case pbcopy:
		args = zero
	default:
		args = zero
	}

	logger.Logger.Printf("getToolArgs(): %s -> %v", tool, args)
	return args
}

func pipTo(tool, text string, args ...string) error {
	logger.Logger.Printf("pipTo(): Executando %s com %d argumentos", tool, len(args))

	cmd := exec.Command(tool, args...)
	logger.Logger.Printf("pipTo(): Comando: %s %s", tool, strings.Join(args, " "))

	input, err := cmd.StdinPipe()
	if err != nil {
		logger.Logger.Printf("pipTo(): ERRO ao criar stdin pipe - %v", err)
		return err
	}

	if err := cmd.Start(); err != nil {
		logger.Logger.Printf("pipTo(): ERRO ao iniciar comando - %v", err)
		return err
	}
	logger.Logger.Printf("pipTo(): Comando iniciado com PID %d", cmd.Process.Pid)

	start := time.Now()
	_, _ = io.WriteString(input, text)
	_ = input.Close()
	err = cmd.Wait()
	duration := time.Since(start)

	if err != nil {
		logger.Logger.Printf("pipTo(): ERRO ao aguardar comando - %v (duração: %v)", err, duration)
	} else {
		logger.Logger.Printf("pipTo(): Comando finalizado com sucesso (duração: %v)", duration)
	}

	return err
}

func hasBinary(name string) bool {
	logger.Logger.Printf("hasBinary(): Verificando se %s existe no PATH", name)
	_, err := exec.LookPath(name)
	exists := err == nil
	logger.Logger.Printf("hasBinary(): %s -> %v", name, exists)
	return exists
}
