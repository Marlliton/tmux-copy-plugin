package app

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/Marlliton/tmux-copy-plugin/internal/clipboard"
	"github.com/Marlliton/tmux-copy-plugin/internal/config"
	"github.com/Marlliton/tmux-copy-plugin/internal/logger"
	"github.com/Marlliton/tmux-copy-plugin/internal/tmux"
)

const (
	colorReset = "\033[0m"
	colorGreen = "\033[32m"
	colorWhite = "\033[97m"
	colorCyan  = "\033[36m"
)

const tmuxPreviousMode = "@go_previous_pane_mode"

func Run(cfg config.Config) error {
	logger.Logger.Printf("run(): Iniciando execução principal")

	if err := tmux.EnsureTmux(); err != nil {
		logger.Logger.Printf("run(): ERRO ensureTmux - %v", err)
		return err
	}
	logger.Logger.Printf("run(): Tmux verificado com sucesso")

	currentMode, err := tmux.GetTmuxMode()
	if err != nil {
		logger.Logger.Printf("run(): ERRO getTmuxMode - %v", err)
		return err
	}
	logger.Logger.Printf("run(): Modo atual do tmux: '%s'", currentMode)

	defer func() {
		logger.Logger.Printf("run(): Definindo opção anterior: %s=%s", tmuxPreviousMode, currentMode)
		_ = tmux.SetTmuxOption(tmuxPreviousMode, currentMode)
	}()

	previous, _ := tmux.GetTmuxOption(tmuxPreviousMode)
	logger.Logger.Printf("run(): Modo anterior: '%s'", previous)

	justLeftCopyMode := strings.HasPrefix(previous, "copy-mode") &&
		!strings.HasPrefix(currentMode, "copy-mode")

	logger.Logger.Printf("run(): justLeftCopyMode = %v (anterior: '%s', atual: '%s')",
		justLeftCopyMode, previous, currentMode)

	if !justLeftCopyMode {
		logger.Logger.Printf("run(): Não saiu do modo de cópia recentemente, saindo")
		return nil
	}

	logger.Logger.Printf("run(): Detectada saída do modo de cópia, processando buffer...")

	text, err := tmux.GetTmuxBuffer()
	if err != nil {
		logger.Logger.Printf("run(): ERRO getTmuxBuffer - %v", err)
		return err
	}
	logger.Logger.Printf("run(): Buffer obtido, tamanho: %d caracteres", len(text))

	clipboardTool, exists := clipboard.HasTool()
	if !exists {
		logger.Logger.Printf("run(): ERRO - Nenhuma ferramenta de clipboard encontrada")
		return errors.New("no clipboard tool found")
	}
	logger.Logger.Printf("run(): Ferramenta de clipboard selecionada: %s", clipboardTool)

	if err := clipboard.Send(clipboardTool, text); err != nil {
		logger.Logger.Printf("run(): ERRO sendToClipboard - %v", err)
		return err
	}
	logger.Logger.Printf("run(): Texto enviado para clipboard com sucesso")

	return showSuccess(cfg, text)
}

func displayPopup(text string) error {
	logger.Logger.Printf("displayPopup(): Exibindo popup, texto tamanho: %d", len(text))
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
		logger.Logger.Printf("displayPopup(): ERRO - %v", err)
	} else {
		logger.Logger.Printf("displayPopup(): Popup exibido com sucesso")
	}

	return err
}

func escapeForShell(s string) string {
	result := strings.ReplaceAll(s, "\"", "\\\"")
	logger.Logger.Printf("escapeForShell(): Entrada: %d chars, Saída: %d chars", len(s), len(result))
	return result
}

func showSuccess(cfg config.Config, text string) error {
	logger.Logger.Printf("showSuccess(): Preparando mensagem de sucesso")
	preview := text
	if len(text) > 350 {
		preview = text[:350] + "..."
		logger.Logger.Printf("showSuccess(): Texto truncado de %d para 350 caracteres", len(text))
	}

	switch cfg.NotificationStyle {
	case config.StylePreview:
		logger.Logger.Printf("showSuccess(): Mensagem preparada, tamanho: %d caracteres", len(preview))
		return displayPopup(preview)
	case config.StyleMessage:
		return exec.Command("tmux", "display-message", "-d", "3000", "✔ Copied text!").Run()
	default:
		return nil
	}
}
