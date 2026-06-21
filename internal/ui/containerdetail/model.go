package containerdetail

import (
	"path"
	"sort"
	"strings"

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
}

func NewModel(
	containerID string,
	containerName string,
	metrics []domain.MetricSample,
	logs []domain.LogLine,
	shellSession *domain.ExecSession,
	files []domain.FileEntry,
) *Model {
	status := shellStateFailed
	if shellSession != nil && len(shellSession.Command) > 0 {
		status = shellStateIdle
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
	}
}

func (m *Model) Init() tea.Cmd {
	return nil
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
		renderTabs(m.tabs, m.activeTab),
		"",
		"Container: " + m.containerName,
		"",
		m.activeContent(),
	}

	return strings.Join(lines, "\n")
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
					"Shell unavailable.",
					"Shell status: " + string(m.shellStatus),
					"Shell error: " + m.shellError,
				}, "\n")
			}
			return "Shell unavailable."
		}

		lines := []string{
			"Shell command: " + strings.Join(m.shellSession.Command, " "),
			"Shell status: " + string(m.shellStatus),
			"Enter Launch Shell",
		}
		if m.shellError != "" {
			lines = append(lines, "Shell error: "+m.shellError)
		}
		return strings.Join(lines, "\n")
	case "Files":
		lines := []string{
			"Path: " + m.currentPath,
		}
		if m.fileError != "" {
			lines = append(lines, "Files error: "+m.fileError)
		}
		if len(m.files) == 0 {
			lines = append(lines, "", "No readable entries.")
			return strings.Join(lines, "\n")
		}

		lines = append(lines, "")
		for index, entry := range m.files {
			prefix := "[f]"
			if entry.IsDir {
				prefix = "[d]"
			}
			cursor := " "
			if index == m.fileCursor {
				cursor = ">"
			}
			lines = append(lines, cursor+" "+prefix+" "+entry.Name+" "+entry.Mode)
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

func renderTabs(tabs []tab, active int) string {
	parts := make([]string, 0, len(tabs))
	for index, tab := range tabs {
		if index == active {
			parts = append(parts, "["+tab.title+"]")
			continue
		}
		parts = append(parts, tab.title)
	}

	return strings.Join(parts, " ")
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
