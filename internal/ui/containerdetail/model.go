package containerdetail

import (
	"path"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/msherburne/dokbox/internal/theme"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/msherburne/dokbox/internal/domain"
)

type tab struct {
	title string
}

type LaunchShellRequest struct {
	ContainerID   string
	ContainerName string
}

type NavigateContainerPathRequest struct {
	ContainerID   string
	ContainerName string
	Path          string
}

type shellState string

const (
	shellStateIdle      shellState = "idle"
	shellStateLaunching shellState = "launching"
	shellStateReady     shellState = "ready"
	shellStateFailed    shellState = "failed"
)

type Model struct {
	containerID   string
	containerName string
	metrics       []domain.MetricSample
	logs          []domain.LogLine
	shellSession  *domain.ExecSession
	files         []domain.FileEntry
	currentPath   string
	fileCursor    int
	fileError     string
	tabs          []tab
	activeTab     int
	shellStatus   shellState
	shellError    string
	width         int
	height        int
	styles        Styles
}

type Styles struct {
	Panel       lipgloss.Style
	Title       lipgloss.Style
	Meta        lipgloss.Style
	Content     lipgloss.Style
	Error       lipgloss.Style
	Info        lipgloss.Style
	ActiveTab   lipgloss.Style
	InactiveTab lipgloss.Style
	SelectedRow lipgloss.Style
}

func NewModel(
	containerID string,
	containerName string,
	metrics []domain.MetricSample,
	logs []domain.LogLine,
	shellSession *domain.ExecSession,
	files []domain.FileEntry,
	styles ...Styles,
) *Model {
	status := shellStateFailed
	if shellSession != nil && len(shellSession.Command) > 0 {
		status = shellStateIdle
	}

	activeStyles := defaultStyles()
	if len(styles) > 0 {
		activeStyles = styles[0]
	}

	return &Model{
		containerID:   containerID,
		containerName: containerName,
		metrics:       append([]domain.MetricSample(nil), metrics...),
		logs:          append([]domain.LogLine(nil), logs...),
		shellSession:  shellSession,
		files:         normalizeFiles("/", files),
		currentPath:   "/",
		shellStatus:   status,
		tabs: []tab{
			{title: "Overview"},
			{title: "Logs"},
			{title: "Shell"},
			{title: "Files"},
		},
		styles: activeStyles,
	}
}

func NewStyles(active theme.Theme) Styles {
	panel := active.Panel
	if panel == "" {
		panel = active.Background
	}

	border := active.Border
	if border == "" {
		border = active.Text
	}

	return Styles{
		Panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(border)).
			Padding(0, 1),
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(active.Text)),
		Meta: lipgloss.NewStyle().
			Foreground(lipgloss.Color(active.Muted)),
		Content: lipgloss.NewStyle().
			Foreground(lipgloss.Color(active.Text)),
		Error: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(active.Error)),
		Info: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(active.Info)),
		ActiveTab: lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Foreground(lipgloss.Color(active.Emphasis)).
			Background(lipgloss.Color(active.Focus)),
		InactiveTab: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color(active.Muted)),
		SelectedRow: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(active.Focus)),
	}
}

func defaultStyles() Styles {
	activeTheme, _ := theme.Resolve(theme.DefaultPreset)
	return NewStyles(activeTheme)
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SetWidth(width int) {
	m.width = width
}

func (m *Model) SetHeight(height int) {
	m.height = height
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "left":
		if m.activeTab > 0 {
			m.activeTab--
		}
	case "right":
		if m.activeTab < len(m.tabs)-1 {
			m.activeTab++
		}
	case "up":
		if m.filesTabActive() && m.fileCursor > 0 {
			m.fileCursor--
		}
	case "down":
		if m.filesTabActive() && m.fileCursor < len(m.files)-1 {
			m.fileCursor++
		}
	case "enter":
		if m.tabs[m.activeTab].title == "Shell" && m.shellSession != nil && len(m.shellSession.Command) > 0 {
			m.shellStatus = shellStateLaunching
			m.shellError = ""
			return m, func() tea.Msg {
				return LaunchShellRequest{
					ContainerID:   m.containerID,
					ContainerName: m.containerName,
				}
			}
		}
		if m.filesTabActive() {
			entry, ok := m.selectedFileEntry()
			if ok && entry.IsDir {
				return m, func() tea.Msg {
					return NavigateContainerPathRequest{
						ContainerID:   m.containerID,
						ContainerName: m.containerName,
						Path:          fileEntryPath(entry, m.currentPath),
					}
				}
			}
		}
	case "backspace":
		if m.filesTabActive() && m.currentPath != "/" {
			return m, func() tea.Msg {
				return NavigateContainerPathRequest{
					ContainerID:   m.containerID,
					ContainerName: m.containerName,
					Path:          parentContainerPath(m.currentPath),
				}
			}
		}
	}

	return m, nil
}

func (m *Model) View() string {
	lines := []string{
		m.styles.Title.Render("Container Detail"),
		m.renderTabs(m.tabs, m.activeTab),
		"",
		m.styles.Meta.Render("Container: " + m.containerName),
		"",
		m.styles.Content.Render(m.activeContent()),
	}
	content := strings.Join(lines, "\n")

	if m.height > 0 {
		maxHeight := m.height - m.styles.Panel.GetVerticalBorderSize()
		content = clipDetailLines(content, maxHeight, m.styles.Info.Render("More content below"))
	}

	panelStyle := m.styles.Panel
	if m.width > 0 {
		width := m.width - panelStyle.GetHorizontalBorderSize()
		if width < 0 {
			width = 0
		}
		panelStyle = panelStyle.Width(width)
	}

	return panelStyle.Render(content)
}

func (m *Model) activeContent() string {
	switch m.tabs[m.activeTab].title {
	case "Logs":
		if len(m.logs) == 0 {
			return "No logs."
		}

		lines := make([]string, 0, len(m.logs))
		for _, line := range m.logs {
			lines = append(lines, line.Text)
		}
		return strings.Join(lines, "\n")
	case "Shell":
		if m.shellSession == nil || len(m.shellSession.Command) == 0 {
			if m.shellError != "" {
				return strings.Join([]string{
					m.styles.Info.Render("Shell unavailable."),
					"Shell status: " + string(m.shellStatus),
					m.styles.Error.Render("Shell error: " + m.shellError),
				}, "\n")
			}
			return m.styles.Info.Render("Shell unavailable.")
		}

		lines := []string{
			"Shell command: " + strings.Join(m.shellSession.Command, " "),
			"Shell status: " + string(m.shellStatus),
			"Enter Launch Shell",
		}
		if m.shellError != "" {
			lines = append(lines, m.styles.Error.Render("Shell error: "+m.shellError))
		}
		return strings.Join(lines, "\n")
	case "Files":
		lines := []string{
			"Path: " + m.currentPath,
		}
		if m.fileError != "" {
			lines = append(lines, m.styles.Error.Render("Files error: "+m.fileError))
		}
		if len(m.files) == 0 {
			lines = append(lines, "", "No readable entries.")
			return strings.Join(lines, "\n")
		}

		lines = append(lines, "", m.renderFilesTable(m.files, m.fileCursor, m.contentWidth()))
		if selected, ok := m.selectedFileEntry(); ok {
			lines = append(lines, "", m.styles.Meta.Render("Selected: "+selected.Name))
		}
		return strings.Join(lines, "\n")
	default:
		if len(m.metrics) == 0 {
			return "No metrics."
		}

		lines := make([]string, 0, len(m.metrics))
		for _, metric := range m.metrics {
			lines = append(lines, metric.Name+": "+metric.Label)
		}
		return strings.Join(lines, "\n")
	}
}

func (m *Model) renderTabs(tabs []tab, active int) string {
	parts := make([]string, 0, len(tabs))
	for index, tab := range tabs {
		if index == active {
			parts = append(parts, m.styles.ActiveTab.Render("["+tab.title+"]"))
			continue
		}
		parts = append(parts, m.styles.InactiveTab.Render(tab.title))
	}

	return strings.Join(parts, " ")
}

func (m *Model) renderFilesTable(files []domain.FileEntry, selected int, maxWidth int) string {
	headers := []string{"Type", "Name", "Mode"}
	widths := []int{4, len("Name"), len("Mode")}

	rows := make([][]string, 0, len(files))
	for _, entry := range files {
		entryType := "file"
		if entry.IsDir {
			entryType = "dir"
		}
		row := []string{entryType, entry.Name, entry.Mode}
		rows = append(rows, row)
		for index, value := range row {
			if len(value) > widths[index] {
				widths[index] = len(value)
			}
		}
	}
	widths = fitTableWidths(widths, maxWidth)

	lines := []string{
		renderTableBorder("top", widths),
		m.renderTableRow(widths, headers, false),
		renderTableBorder("middle", widths),
	}
	for index, row := range rows {
		lines = append(lines, m.renderTableRow(widths, row, index == selected))
	}
	lines = append(lines, renderTableBorder("bottom", widths))
	return strings.Join(lines, "\n")
}

func renderTableBorder(position string, widths []int) string {
	left, middle, right := "┌", "┬", "┐"
	switch position {
	case "middle":
		left, middle, right = "├", "┼", "┤"
	case "bottom":
		left, middle, right = "└", "┴", "┘"
	}

	parts := make([]string, 0, len(widths))
	for _, width := range widths {
		parts = append(parts, strings.Repeat("─", width+2))
	}
	return left + strings.Join(parts, middle) + right
}

func (m *Model) renderTableRow(widths []int, values []string, selected bool) string {
	cells := make([]string, 0, len(values))
	for index, value := range values {
		cell := lipgloss.NewStyle().Width(widths[index]).Render(truncateTableValue(value, widths[index]))
		cells = append(cells, " "+cell+" ")
	}

	row := "│" + strings.Join(cells, "│") + "│"
	if selected {
		return m.styles.SelectedRow.Render(row)
	}
	return row
}

func (m *Model) ApplyShellLaunchResult(session *domain.ExecSession, err error) {
	if err != nil {
		m.shellStatus = shellStateFailed
		m.shellError = err.Error()
		return
	}

	if session != nil {
		m.shellSession = session
	}
	m.shellStatus = shellStateReady
	m.shellError = ""
}

func (m *Model) ApplyFileNavigationResult(currentPath string, files []domain.FileEntry, err error) {
	if err != nil {
		m.fileError = err.Error()
		return
	}

	m.currentPath = normalizeContainerPath(currentPath)
	m.files = normalizeFiles(m.currentPath, files)
	m.fileCursor = 0
	m.fileError = ""
}

func (m *Model) filesTabActive() bool {
	return m.tabs[m.activeTab].title == "Files"
}

func (m *Model) selectedFileEntry() (domain.FileEntry, bool) {
	if len(m.files) == 0 || m.fileCursor < 0 || m.fileCursor >= len(m.files) {
		return domain.FileEntry{}, false
	}

	return m.files[m.fileCursor], true
}

func fileEntryPath(entry domain.FileEntry, currentPath string) string {
	if entry.Path != "" {
		return normalizeContainerPath(entry.Path)
	}
	return normalizeContainerPath(path.Join(currentPath, entry.Name))
}

func parentContainerPath(currentPath string) string {
	currentPath = normalizeContainerPath(currentPath)
	if currentPath == "/" {
		return currentPath
	}

	parent := path.Dir(currentPath)
	if parent == "." {
		return "/"
	}
	return normalizeContainerPath(parent)
}

func normalizeContainerPath(value string) string {
	if value == "" {
		return "/"
	}

	cleaned := path.Clean(value)
	if cleaned == "." {
		return "/"
	}
	if !strings.HasPrefix(cleaned, "/") {
		return "/" + cleaned
	}
	return cleaned
}

func normalizeFiles(currentPath string, files []domain.FileEntry) []domain.FileEntry {
	normalized := make([]domain.FileEntry, 0, len(files))
	for _, entry := range files {
		entryPath := entry.Path
		if entryPath == "" {
			entryPath = path.Join(currentPath, entry.Name)
		}
		normalized = append(normalized, domain.FileEntry{
			Path:  normalizeContainerPath(entryPath),
			Name:  entry.Name,
			IsDir: entry.IsDir,
			Size:  entry.Size,
			Mode:  entry.Mode,
		})
	}

	sort.SliceStable(normalized, func(i, j int) bool {
		if normalized[i].IsDir != normalized[j].IsDir {
			return normalized[i].IsDir
		}
		return normalized[i].Name < normalized[j].Name
	})

	return normalized
}

func (m *Model) contentWidth() int {
	if m.width <= 0 {
		return 0
	}
	return m.width - m.styles.Panel.GetHorizontalBorderSize()
}

func fitTableWidths(widths []int, maxWidth int) []int {
	if maxWidth <= 0 {
		return widths
	}

	available := maxWidth - (len(widths)*3 + 1)
	if available <= 0 {
		out := make([]int, len(widths))
		for i := range out {
			out[i] = 1
		}
		return out
	}

	out := append([]int(nil), widths...)
	minWidth := 4
	for intsTotal(out) > available {
		shrunk := false
		for i := range out {
			if out[i] > minWidth && intsTotal(out) > available {
				out[i]--
				shrunk = true
			}
		}
		if !shrunk {
			break
		}
	}

	return out
}

func intsTotal(values []int) int {
	sum := 0
	for _, value := range values {
		sum += value
	}
	return sum
}

func truncateTableValue(value string, width int) string {
	if width <= 0 {
		return ""
	}

	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width == 1 {
		return "…"
	}
	return string(runes[:width-1]) + "…"
}

func clipDetailLines(content string, maxHeight int, indicator string) string {
	if maxHeight <= 0 {
		return indicator
	}

	lines := strings.Split(content, "\n")
	if len(lines) <= maxHeight {
		return content
	}
	if maxHeight == 1 {
		return indicator
	}

	clipped := append([]string{}, lines[:maxHeight-1]...)
	clipped = append(clipped, indicator)
	return strings.Join(clipped, "\n")
}
