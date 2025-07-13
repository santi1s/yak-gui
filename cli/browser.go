package cli

import (
	"os/exec"
)

type Browser interface {
	Go(string) error
}

type WinBrowser struct {
	cmd *exec.Cmd
}

type MacosBrowser struct {
	cmd *exec.Cmd
}

type LinuxBrowser struct {
	cmd *exec.Cmd
}

func (w *WinBrowser) Go(url string) error {
	w.cmd = exec.Command("cmd", "/c", "start", url)
	return w.cmd.Start()
}

func (m *MacosBrowser) Go(url string) error {
	m.cmd = exec.Command("open", url)
	return m.cmd.Start()
}

func (l *LinuxBrowser) Go(url string) error {
	l.cmd = exec.Command("xdg-open", url)
	return l.cmd.Start()
}
