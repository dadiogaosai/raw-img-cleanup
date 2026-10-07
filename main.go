package main

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/dadiogaosai/rawtidy/cleanup"
)

type cleanupFinished struct {
	result cleanup.Result
	err    error
}

type app struct {
	config   config
	picker   filepicker.Model
	input    textinput.Model
	jpeg     string
	raw      string
	stage    int // 0: JPEG, 1: RAW, 2: running, 3: result
	editing  bool
	err      error
	fatalErr error
	result   cleanup.Result
	height   int
	width    int
	scroll   int
	details  []string
}

func newPicker(path string) filepicker.Model {
	picker := filepicker.New()
	picker.CurrentDirectory = path
	picker.FileAllowed = false
	picker.DirAllowed = false // Enter navigates; s selects the current folder.
	picker.AutoHeight = false
	picker.SetHeight(12)
	return picker
}

func newApp(cfg config) (app, error) {
	path, err := cfg.startDirectory(cfg.LastJPEG)
	if err != nil {
		return app{}, err
	}
	input := textinput.New()
	input.Placeholder = "Type or paste a directory path"
	input.CharLimit = 4096
	return app{config: cfg, picker: newPicker(path), input: input, height: 24, width: 80}, nil
}

func (a app) Init() tea.Cmd { return a.picker.Init() }

func (a app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		a.height = size.Height
		a.width = size.Width
	}
	if key, ok := msg.(tea.KeyPressMsg); ok && key.String() == "ctrl+c" {
		return a, tea.Quit
	}
	if finished, ok := msg.(cleanupFinished); ok {
		a.stage = 3
		a.result = finished.result
		a.err = finished.err
		for _, moved := range a.result.Files {
			a.details = append(a.details, fmt.Sprintf("Moved: %s -> %s", moved.Source, moved.Destination))
		}
		for _, failure := range a.result.Failures {
			a.details = append(a.details, fmt.Sprintf("Failed: %s: %v", failure.Source, failure.Err))
		}
		return a, nil
	}
	if a.stage == 2 {
		return a, nil
	}
	if a.stage == 3 {
		rows := a.resultRows()
		if key, ok := msg.(tea.KeyPressMsg); ok {
			switch key.String() {
			case "q":
				return a, tea.Quit
			case "down", "j":
				a.scroll++
			case "up", "k":
				a.scroll--
			case "pgdown":
				a.scroll += a.resultPageSize()
			case "pgup":
				a.scroll -= a.resultPageSize()
			}
		}
		a.scroll = max(0, min(a.scroll, max(0, len(rows)-a.resultPageSize())))
		return a, nil
	}
	if key, ok := msg.(tea.KeyPressMsg); ok {
		if a.editing {
			switch key.String() {
			case "esc":
				a.editing = false
				a.input.Blur()
				return a, nil
			case "enter":
				return a.selectDirectory(a.input.Value())
			}
		} else {
			switch key.String() {
			case "q":
				return a, tea.Quit
			case "e":
				a.editing = true
				a.input.SetValue(a.picker.CurrentDirectory)
				return a, a.input.Focus()
			case "s":
				return a.selectDirectory(a.picker.HighlightedPath())
			case ".":
				return a.selectDirectory(a.picker.CurrentDirectory)
			}
		}
	}
	if a.editing {
		var cmd tea.Cmd
		a.input, cmd = a.input.Update(msg)
		return a, cmd
	}
	var cmd tea.Cmd
	a.picker, cmd = a.picker.Update(msg)
	return a, cmd
}

func (a app) selectDirectory(path string) (tea.Model, tea.Cmd) {
	path = strings.TrimSpace(path)
	if a.stage == 0 {
		jpeg, err := cleanup.CheckDirectory(path)
		if err != nil {
			a.err = fmt.Errorf("JPEG directory: %w", err)
			return a, nil
		}
		rawStart, err := a.config.startDirectory(a.config.LastRAW)
		if err != nil {
			a.err = fmt.Errorf("RAW starting directory: %w", err)
			return a, nil
		}
		a.config.LastJPEG = jpeg
		if err := a.config.save(); err != nil {
			a.fatalErr = fmt.Errorf("save JPEG directory: %w", err)
			return a, tea.Quit
		}
		a.jpeg = jpeg
		a.picker = newPicker(rawStart)
		a.stage = 1
		a.err = nil
		a.editing = false
		a.input.Blur()
		return a, a.picker.Init()
	}
	paths, err := cleanup.Validate(a.jpeg, path)
	if err != nil {
		a.err = err
		return a, nil
	}
	a.config.LastRAW = paths.RAW
	if err := a.config.save(); err != nil {
		a.fatalErr = fmt.Errorf("save RAW directory: %w", err)
		return a, tea.Quit
	}
	a.raw = paths.RAW
	a.err = nil
	a.stage = 2
	a.editing = false
	a.input.Blur()
	jpeg, raw := a.jpeg, a.raw
	return a, func() tea.Msg {
		result, err := cleanup.Run(jpeg, raw)
		return cleanupFinished{result: result, err: err}
	}
}

func (a app) View() tea.View {
	var b strings.Builder
	b.WriteString("RAW cleanup\n\n")
	switch a.stage {
	case 0, 1:
		if a.stage == 0 {
			b.WriteString("Choose JPEG directory\n")
		} else {
			fmt.Fprintf(&b, "JPEG: %s\nChoose RAW directory\n", a.jpeg)
		}
		b.WriteString("JPEG stems match RAW stems anywhere in the selected trees.\n\n")
		if a.editing {
			b.WriteString("Enter a path, then press Enter. Esc returns to browsing.\n")
			b.WriteString(a.input.View())
		} else {
			fmt.Fprintf(&b, "Browsing: %s\n", a.picker.CurrentDirectory)
			b.WriteString("Enter/right: open folder · h/left: parent · s: select highlighted folder · .: select current folder · e: enter path · q: quit\n\n")
			b.WriteString(a.picker.View())
		}
		if a.err != nil {
			fmt.Fprintf(&b, "\nError: %v\n", a.err)
		}
	case 2:
		fmt.Fprintf(&b, "JPEG: %s\nRAW: %s\n\nScanning and moving unmatched RAW files...\n", a.jpeg, a.raw)
	case 3:
		if a.err != nil {
			fmt.Fprintf(&b, "Error: %v\n", a.err)
		} else {
			for _, line := range a.wrapRows(fmt.Sprintf("Kept: %d · Moved: %d · Failed: %d", a.result.Kept, a.result.Moved, a.result.Failed)) {
				b.WriteString(line + "\n")
			}
			if a.result.Moved == 0 && a.result.Failed == 0 {
				for _, line := range a.wrapRows("Nothing was moved.") {
					b.WriteString(line + "\n")
				}
			}
			rows := a.resultRows()
			if len(rows) > 0 {
				start := max(0, min(a.scroll, len(rows)-1))
				end := min(len(rows), start+a.resultPageSize())
				for _, line := range rows[start:end] {
					b.WriteString(line + "\n")
				}
			}
		}
		b.WriteString("\nj/k scroll, q quit\n")
	}
	return tea.NewView(b.String())
}

func (a app) resultPageSize() int {
	summary := fmt.Sprintf("Kept: %d · Moved: %d · Failed: %d", a.result.Kept, a.result.Moved, a.result.Failed)
	reserved := 2 + len(a.wrapRows(summary)) + 2 // title, blank line, footer
	if a.result.Moved == 0 && a.result.Failed == 0 {
		reserved += len(a.wrapRows("Nothing was moved."))
	}
	return max(1, a.height-reserved)
}

func (a app) resultRows() []string {
	var rows []string
	for _, detail := range a.details {
		rows = append(rows, a.wrapRows(detail)...)
	}
	return rows
}

func (a app) wrapRows(text string) []string {
	return strings.Split(ansi.Hardwrap(text, max(1, a.width-1), true), "\n")
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	cfg, err := loadConfig(home)
	if err != nil {
		return err
	}
	a, err := newApp(cfg)
	if err != nil {
		return err
	}
	model, err := tea.NewProgram(a).Run()
	if err != nil {
		return err
	}
	return model.(app).fatalErr
}
