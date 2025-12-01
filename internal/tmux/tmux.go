package tmux

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/Marlliton/tmux-copy-plugin/internal/logger"
)

func EnsureTmux() error {
	logger.Logger.Printf("ensureTmux(): Verificando se tmux está disponível")
	_, err := exec.LookPath("tmux")
	if err != nil {
		logger.Logger.Printf("ensureTmux(): ERRO - tmux não encontrado")
		return errors.New("tmux not found")
	}
	logger.Logger.Printf("ensureTmux(): Tmux encontrado")
	return nil
}

func GetTmuxMode() (string, error) {
	logger.Logger.Printf("getTmuxMode(): Executando comando tmux display-message")
	cmd := exec.Command("tmux", "display-message", "-p", "#{pane_mode}")
	out, err := cmd.Output()
	if err != nil {
		logger.Logger.Printf("getTmuxMode(): ERRO - %v", err)
		return "", err
	}
	result := strings.TrimSpace(string(out))
	logger.Logger.Printf("getTmuxMode(): Resultado: '%s'", result)
	return result, nil
}

func GetTmuxOption(option string) (string, error) {
	logger.Logger.Printf("getTmuxOption(): Obtendo opção: %s", option)
	cmd := exec.Command("tmux", "show-options", "-gqv", option)
	out, err := cmd.Output()
	if err != nil {
		logger.Logger.Printf("getTmuxOption(): ERRO - %v", err)
		return "", nil
	}
	result := strings.TrimSpace(string(out))
	logger.Logger.Printf("getTmuxOption(): %s = '%s'", option, result)
	return result, nil
}

func SetTmuxOption(option, value string) error {
	logger.Logger.Printf("setTmuxOption(): Definindo %s = '%s'", option, value)
	cmd := exec.Command("tmux", "set-option", "-g", option, value)
	err := cmd.Run()
	if err != nil {
		logger.Logger.Printf("setTmuxOption(): ERRO - %v", err)
	} else {
		logger.Logger.Printf("setTmuxOption(): Sucesso")
	}
	return err
}

func GetTmuxBuffer() (string, error) {
	logger.Logger.Printf("getTmuxBuffer(): Obtendo buffer do tmux")
	cmd := exec.Command("tmux", "show-buffer")
	out, err := cmd.Output()
	if err != nil {
		logger.Logger.Printf("getTmuxBuffer(): ERRO - %v", err)
		return "", errors.New("there is nothing in the buffer")
	}
	text := string(out)
	if text == "" {
		logger.Logger.Printf("getTmuxBuffer(): ERRO - Buffer vazio")
		return "", errors.New("no text selected")
	}

	logger.Logger.Printf("getTmuxBuffer(): Buffer obtido, primeiros 100 chars: '%s'",
		truncateText(text, 100))
	return text, nil
}

func truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}
