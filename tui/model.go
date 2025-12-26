package tui

import (
	"encoding/base64"
	"fmt"
	"net/netip"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Model is the Bubble Tea model for the TUI.
type Model struct {
	root           *SubnetNode
	rows           []*SubnetNode // Flattened list of visible leaf nodes
	cursor         int
	width          int
	height         int
	maxSplitDepth  int // Maximum prefix length (e.g., 30 for /30)
	initialPrefix  int // The starting prefix length
	scrollOffset   int // Horizontal scroll offset for split columns
	verticalScroll int // Vertical scroll offset for rows
	help           help.Model
	keys           keyMap
	statusMsg      string // Status message to display
}

// NewModel creates a new TUI model from a CIDR string.
// Optional targetBits parameter specifies initial split depth (0 means no initial split).
func NewModel(cidr string, targetBits int) (Model, error) {
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return Model{}, fmt.Errorf("invalid CIDR: %w", err)
	}

	// Normalize to network address
	prefix = netip.PrefixFrom(prefix.Masked().Addr(), prefix.Bits())

	// Validate targetBits if specified
	if targetBits > 0 {
		if targetBits <= prefix.Bits() {
			return Model{}, fmt.Errorf("target subnet size /%d must be larger than the network size /%d", targetBits, prefix.Bits())
		}
		if targetBits > 30 {
			return Model{}, fmt.Errorf("target subnet size /%d exceeds maximum allowed /30", targetBits)
		}
	}

	root := createSubnetNode(prefix, nil)

	// Pre-split to target depth if specified
	if targetBits > 0 {
		root.SplitToDepth(targetBits)
	}

	m := Model{
		root:          root,
		cursor:        0,
		maxSplitDepth: 30, // Allow splitting down to /30
		initialPrefix: prefix.Bits(),
		scrollOffset:  0,
		help:          help.New(),
		keys:          defaultKeys,
	}
	m.updateRows()

	return m, nil
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// clearStatusMsg is a message type to clear the status.
type clearStatusMsg struct{}

// clearStatusAfter returns a command that clears status after delay.
func clearStatusAfter() tea.Cmd {
	return tea.Tick(time.Second*3, func(_ time.Time) tea.Msg {
		return clearStatusMsg{}
	})
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case clearStatusMsg:
		m.statusMsg = ""
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
	}

	return m, nil
}

// handleKeyPress handles keyboard input.
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}

	case key.Matches(msg, m.keys.Down):
		if m.cursor < len(m.rows)-1 {
			m.cursor++
		}

	case key.Matches(msg, m.keys.PageUp):
		viewportHeight := max(3, m.height-10)
		m.verticalScroll = max(0, m.verticalScroll-viewportHeight)
		m.cursor = max(0, m.cursor-viewportHeight)

	case key.Matches(msg, m.keys.PageDown):
		viewportHeight := max(3, m.height-10)
		m.verticalScroll += viewportHeight
		maxScroll := max(0, len(m.rows)-viewportHeight)
		m.verticalScroll = min(m.verticalScroll, maxScroll)
		m.cursor = min(m.cursor+viewportHeight, len(m.rows)-1)

	case key.Matches(msg, m.keys.Left):
		if m.scrollOffset > 0 {
			m.scrollOffset--
		}

	case key.Matches(msg, m.keys.Right):
		maxBits := m.getMaxBits()
		numCols := maxBits - m.initialPrefix + 1
		if m.scrollOffset < numCols-1 {
			m.scrollOffset++
		}

	case key.Matches(msg, m.keys.Split):
		if len(m.rows) > 0 && m.cursor < len(m.rows) {
			node := m.rows[m.cursor]
			if node.CIDR.Bits() < m.maxSplitDepth {
				node.Split()
				m.updateRows()
			}
		}

	case key.Matches(msg, m.keys.Join):
		if len(m.rows) > 0 && m.cursor < len(m.rows) {
			node := m.rows[m.cursor]
			if node.Parent != nil {
				node.Parent.Join()
				m.updateRows()
			}
		}

	case key.Matches(msg, m.keys.Export):
		m.statusMsg = "Press 'q' to quit and see JSON output"
		return m, clearStatusAfter()

	case key.Matches(msg, m.keys.Copy):
		jsonStr := m.exportJSON()
		encoded := base64.StdEncoding.EncodeToString([]byte(jsonStr))
		fmt.Printf("\033]52;c;%s\a", encoded)
		m.statusMsg = "✓ Copied to clipboard!"
		return m, clearStatusAfter()

	case key.Matches(msg, m.keys.Help):
		m.help.ShowAll = !m.help.ShowAll
	}

	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Title
	title := titleStyle.Render(fmt.Sprintf("🌐 Subnet Calculator - %s", m.root.CIDR.String()))
	b.WriteString(title)
	b.WriteString("\n\n")

	// Build the table
	table := m.renderTable()
	b.WriteString(table)
	b.WriteString("\n")

	// Status message
	if m.statusMsg != "" {
		b.WriteString(statusStyle.Render(m.statusMsg))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Help
	b.WriteString(m.help.View(m.keys))

	return b.String()
}

// updateRows flattens the tree into visible leaf nodes.
func (m *Model) updateRows() {
	m.rows = nil
	collectLeaves(m.root, &m.rows)

	// Adjust cursor to valid bounds
	m.cursor = max(0, min(m.cursor, len(m.rows)-1))
}

// hasSplits returns true if any subnet has been split.
func (m *Model) hasSplits() bool {
	return m.root.IsSplit
}

// getMaxBits returns the maximum prefix bits of any leaf node.
func (m *Model) getMaxBits() int {
	maxBits := m.initialPrefix
	for _, row := range m.rows {
		if row.CIDR.Bits() > maxBits {
			maxBits = row.CIDR.Bits()
		}
	}
	return maxBits
}

// exportJSON returns the current state as JSON.
func (m *Model) exportJSON() string {
	json, err := m.root.ExportJSON()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return json
}

// calculateColumnWidths determines optimal widths based on content and terminal width.
func (m *Model) calculateColumnWidths() columnWidths {
	minWidths := minColumnWidths()

	// Determine if IPv6 by checking root address
	isIPv6 := m.root.CIDR.Addr().Is6()

	// Calculate content-based widths
	var maxSubnet, maxMask, maxRange, maxHosts int

	for _, node := range m.rows {
		cidrLen := len(node.CIDR.String())
		if cidrLen > maxSubnet {
			maxSubnet = cidrLen
		}

		maskLen := len(node.SubnetMask.String())
		if maskLen > maxMask {
			maxMask = maskLen
		}

		networkAddr := node.CIDR.Masked().Addr()
		rangeStr := formatRangeAbbreviated(node.FirstIP.String(), node.LastIP.String(), networkAddr.String())
		rangeLen := len(rangeStr)
		if rangeLen > maxRange {
			maxRange = rangeLen
		}

		hostsStr := formatNumber(node.Hosts)
		hostsLen := len(hostsStr)
		if hostsLen > maxHosts {
			maxHosts = hostsLen
		}
	}

	// Add padding (2 chars for spacing)
	maxSubnet += 2
	maxMask += 2
	maxRange += 2
	maxHosts += 2

	// Apply minimums
	maxSubnet = max(maxSubnet, minWidths.subnet)
	maxMask = max(maxMask, minWidths.mask)
	maxRange = max(maxRange, minWidths.rangeCol)
	maxHosts = max(maxHosts, minWidths.hosts)

	// Calculate split column width
	splitColWidth := minWidths.splitCol
	if isIPv6 {
		splitColWidth = 6 // /xxx for IPv6 prefixes up to /128
	}

	// Calculate total needed width
	mainWidth := maxSubnet + maxMask + maxRange + maxHosts + 8
	hasSplits := m.hasSplits()
	maxBits := m.getMaxBits()
	numSplitLevels := 0
	if hasSplits {
		numSplitLevels = maxBits - m.initialPrefix + 1
	}
	splitWidth := numSplitLevels * splitColWidth

	totalNeeded := mainWidth + splitWidth

	// If terminal is wide enough, use calculated widths
	if totalNeeded <= m.width || m.width == 0 {
		return columnWidths{
			subnet:   maxSubnet,
			mask:     maxMask,
			rangeCol: maxRange,
			hosts:    maxHosts,
			splitCol: splitColWidth,
		}
	}

	// Terminal is too narrow - need to shrink columns proportionally
	availableMain := m.width - splitWidth - 8
	minTotal := minWidths.subnet + minWidths.mask + minWidths.rangeCol + minWidths.hosts
	if availableMain < minTotal {
		return columnWidths{
			subnet:   minWidths.subnet,
			mask:     minWidths.mask,
			rangeCol: minWidths.rangeCol,
			hosts:    minWidths.hosts,
			splitCol: splitColWidth,
		}
	}

	// Distribute available space proportionally but respect minimums
	totalContent := maxSubnet + maxMask + maxRange + maxHosts
	scale := float64(availableMain) / float64(totalContent)

	subnetW := max(int(float64(maxSubnet)*scale), minWidths.subnet)
	maskW := max(int(float64(maxMask)*scale), minWidths.mask)
	rangeW := max(int(float64(maxRange)*scale), minWidths.rangeCol)
	hostsW := max(int(float64(maxHosts)*scale), minWidths.hosts)

	return columnWidths{
		subnet:   subnetW,
		mask:     maskW,
		rangeCol: rangeW,
		hosts:    hostsW,
		splitCol: splitColWidth,
	}
}

// Run starts the TUI.
// Optional initialSplit parameter specifies initial split depth (0 means no initial split).
func Run(cidr string, initialSplit int) error {
	model, err := NewModel(cidr, initialSplit)
	if err != nil {
		return err
	}

	// Don't use alt screen so the final state is preserved when quitting
	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Print a newline to separate from the TUI output
	fmt.Println()

	// If user requested export, print JSON
	if m, ok := finalModel.(Model); ok {
		if m.statusMsg == "Press 'q' to quit and see JSON output" {
			fmt.Println(m.exportJSON())
		}
	}

	return nil
}
