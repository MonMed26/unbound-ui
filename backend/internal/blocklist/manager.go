package blocklist

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Source represents a blocklist source
type Source struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Enabled   bool      `json:"enabled"`
	Count     int       `json:"count"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Manager handles DNS blocklist operations
type Manager struct {
	mu          sync.RWMutex
	dataDir     string
	sources     []Source
	whitelist   map[string]bool
	manualBlock map[string]bool
	outputPath  string
}

func NewManager(dataDir, outputPath string) *Manager {
	if dataDir == "" {
		dataDir = "/var/lib/unbound-ui/blocklist"
	}
	if outputPath == "" {
		outputPath = "/etc/unbound/unbound.conf.d/blocklist.conf"
	}

	m := &Manager{
		dataDir:     dataDir,
		outputPath:  outputPath,
		whitelist:   make(map[string]bool),
		manualBlock: make(map[string]bool),
	}

	os.MkdirAll(dataDir, 0755)
	m.loadState()
	return m
}

// GetSources returns all blocklist sources
func (m *Manager) GetSources() []Source {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sources
}

// AddSource adds a new blocklist source
func (m *Manager) AddSource(name, url string) (*Source, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	source := Source{
		ID:      fmt.Sprintf("src_%d", time.Now().UnixNano()),
		Name:    name,
		URL:     url,
		Enabled: true,
	}
	m.sources = append(m.sources, source)
	m.saveState()
	return &source, nil
}

// RemoveSource removes a blocklist source
func (m *Manager) RemoveSource(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, s := range m.sources {
		if s.ID == id {
			m.sources = append(m.sources[:i], m.sources[i+1:]...)
			m.saveState()
			return nil
		}
	}
	return fmt.Errorf("source not found: %s", id)
}

// ToggleSource enables/disables a source
func (m *Manager) ToggleSource(id string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, s := range m.sources {
		if s.ID == id {
			m.sources[i].Enabled = enabled
			m.saveState()
			return nil
		}
	}
	return fmt.Errorf("source not found: %s", id)
}

// BlockDomain manually blocks a domain
func (m *Manager) BlockDomain(domain string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	domain = strings.ToLower(strings.TrimSpace(domain))
	m.manualBlock[domain] = true
	delete(m.whitelist, domain)
	m.saveState()
}

// UnblockDomain removes a domain from manual block
func (m *Manager) UnblockDomain(domain string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	domain = strings.ToLower(strings.TrimSpace(domain))
	delete(m.manualBlock, domain)
	m.saveState()
}

// WhitelistDomain adds a domain to whitelist
func (m *Manager) WhitelistDomain(domain string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	domain = strings.ToLower(strings.TrimSpace(domain))
	m.whitelist[domain] = true
	delete(m.manualBlock, domain)
	m.saveState()
}

// RemoveWhitelist removes a domain from whitelist
func (m *Manager) RemoveWhitelist(domain string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	domain = strings.ToLower(strings.TrimSpace(domain))
	delete(m.whitelist, domain)
	m.saveState()
}

// GetWhitelist returns all whitelisted domains
func (m *Manager) GetWhitelist() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var domains []string
	for d := range m.whitelist {
		domains = append(domains, d)
	}
	return domains
}

// GetManualBlocks returns manually blocked domains
func (m *Manager) GetManualBlocks() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var domains []string
	for d := range m.manualBlock {
		domains = append(domains, d)
	}
	return domains
}

// Update downloads all enabled sources and regenerates the blocklist config
func (m *Manager) Update() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	allDomains := make(map[string]bool)

	// Download each enabled source
	for i, source := range m.sources {
		if !source.Enabled {
			continue
		}

		domains, err := m.downloadSource(source.URL)
		if err != nil {
			continue // Skip failed sources
		}

		m.sources[i].Count = len(domains)
		m.sources[i].UpdatedAt = time.Now()

		for _, d := range domains {
			allDomains[d] = true
		}
	}

	// Add manual blocks
	for d := range m.manualBlock {
		allDomains[d] = true
	}

	// Remove whitelisted domains
	for d := range m.whitelist {
		delete(allDomains, d)
	}

	// Generate unbound config
	if err := m.generateConfig(allDomains); err != nil {
		return err
	}

	m.saveState()
	return nil
}

// GetStats returns blocklist statistics
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalDomains := 0
	for _, s := range m.sources {
		if s.Enabled {
			totalDomains += s.Count
		}
	}
	totalDomains += len(m.manualBlock)

	return map[string]interface{}{
		"total_sources":     len(m.sources),
		"enabled_sources":   countEnabled(m.sources),
		"total_domains":     totalDomains,
		"manual_blocks":     len(m.manualBlock),
		"whitelisted":       len(m.whitelist),
	}
}

func (m *Manager) downloadSource(url string) ([]string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return parseDomainList(resp.Body), nil
}

func parseDomainList(r io.Reader) []string {
	var domains []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}

		// Handle hosts file format: "0.0.0.0 domain.com" or "127.0.0.1 domain.com"
		if strings.HasPrefix(line, "0.0.0.0") || strings.HasPrefix(line, "127.0.0.1") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				domain := strings.ToLower(parts[1])
				if isValidDomain(domain) {
					domains = append(domains, domain)
				}
			}
			continue
		}

		// Handle plain domain format
		domain := strings.ToLower(strings.Fields(line)[0])
		if isValidDomain(domain) {
			domains = append(domains, domain)
		}
	}
	return domains
}

func isValidDomain(domain string) bool {
	if domain == "" || domain == "localhost" || domain == "localhost.localdomain" {
		return false
	}
	if !strings.Contains(domain, ".") {
		return false
	}
	return true
}

func (m *Manager) generateConfig(domains map[string]bool) error {
	dir := filepath.Dir(m.outputPath)
	os.MkdirAll(dir, 0755)

	f, err := os.Create(m.outputPath)
	if err != nil {
		return fmt.Errorf("failed to create blocklist config: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	w.WriteString("# Auto-generated by Unbound UI - DO NOT EDIT MANUALLY\n")
	w.WriteString(fmt.Sprintf("# Generated at: %s\n", time.Now().Format(time.RFC3339)))
	w.WriteString(fmt.Sprintf("# Total blocked domains: %d\n\n", len(domains)))
	w.WriteString("server:\n")

	for domain := range domains {
		w.WriteString(fmt.Sprintf("    local-zone: \"%s\" always_refuse\n", domain))
	}

	return w.Flush()
}

type state struct {
	Sources     []Source          `json:"sources"`
	Whitelist   map[string]bool   `json:"whitelist"`
	ManualBlock map[string]bool   `json:"manual_block"`
}

func (m *Manager) saveState() {
	s := state{
		Sources:     m.sources,
		Whitelist:   m.whitelist,
		ManualBlock: m.manualBlock,
	}
	data, _ := json.MarshalIndent(s, "", "  ")
	os.WriteFile(filepath.Join(m.dataDir, "state.json"), data, 0644)
}

func (m *Manager) loadState() {
	data, err := os.ReadFile(filepath.Join(m.dataDir, "state.json"))
	if err != nil {
		return
	}
	var s state
	if err := json.Unmarshal(data, &s); err != nil {
		return
	}
	m.sources = s.Sources
	if s.Whitelist != nil {
		m.whitelist = s.Whitelist
	}
	if s.ManualBlock != nil {
		m.manualBlock = s.ManualBlock
	}
}

func countEnabled(sources []Source) int {
	count := 0
	for _, s := range sources {
		if s.Enabled {
			count++
		}
	}
	return count
}
