package unbound

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ConfigManager handles reading and writing unbound.conf
type ConfigManager struct {
	ConfigPath string
}

func NewConfigManager(configPath string) *ConfigManager {
	if configPath == "" {
		configPath = "/etc/unbound/unbound.conf"
	}
	return &ConfigManager{ConfigPath: configPath}
}

// ServerConfig represents the server section of unbound.conf
type ServerConfig struct {
	Interface     []string `json:"interface"`
	Port          int      `json:"port"`
	AccessControl []string `json:"access_control"`
	DoIPv4        bool     `json:"do_ipv4"`
	DoIPv6        bool     `json:"do_ipv6"`
	DoUDP         bool     `json:"do_udp"`
	DoTCP         bool     `json:"do_tcp"`
	Verbosity     int      `json:"verbosity"`
	NumThreads    int      `json:"num_threads"`
	CacheMinTTL   int      `json:"cache_min_ttl"`
	CacheMaxTTL   int      `json:"cache_max_ttl"`
	Prefetch      bool     `json:"prefetch"`
	HideIdentity  bool     `json:"hide_identity"`
	HideVersion   bool     `json:"hide_version"`
}

// ForwardZone represents a forward-zone section
type ForwardZone struct {
	Name    string   `json:"name"`
	Addrs   []string `json:"forward_addr"`
	TLSUp   string   `json:"forward_tls_upstream,omitempty"`
}

// UnboundConfig represents the full parsed config
type UnboundConfig struct {
	Raw          string        `json:"raw"`
	Server       *ServerConfig `json:"server,omitempty"`
	ForwardZones []ForwardZone `json:"forward_zones,omitempty"`
	Includes     []string      `json:"includes,omitempty"`
}

// ReadConfig reads and parses the unbound configuration file
func (cm *ConfigManager) ReadConfig() (*UnboundConfig, error) {
	data, err := os.ReadFile(cm.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	config := &UnboundConfig{
		Raw: string(data),
	}

	config.Server = cm.parseServerSection(string(data))
	config.ForwardZones = cm.parseForwardZones(string(data))
	config.Includes = cm.parseIncludes(string(data))

	return config, nil
}

// WriteConfig writes the raw config to file with backup
func (cm *ConfigManager) WriteConfig(raw string) error {
	// Create backup
	if err := cm.backup(); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	if err := os.WriteFile(cm.ConfigPath, []byte(raw), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}
	return nil
}

// ValidateConfig runs unbound-checkconf on the config
func (cm *ConfigManager) ValidateConfig() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "unbound-checkconf", cm.ConfigPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", strings.TrimSpace(string(output)))
	}
	return nil
}

// backup creates a timestamped backup of the current config
func (cm *ConfigManager) backup() error {
	data, err := os.ReadFile(cm.ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No file to backup
		}
		return err
	}

	dir := filepath.Dir(cm.ConfigPath)
	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(dir, fmt.Sprintf("unbound.conf.bak.%s", timestamp))

	return os.WriteFile(backupPath, data, 0644)
}

func (cm *ConfigManager) parseServerSection(content string) *ServerConfig {
	cfg := &ServerConfig{
		Port:     53,
		DoIPv4:   true,
		DoUDP:    true,
		DoTCP:    true,
		DoIPv6:   true,
		Prefetch: false,
	}

	inServer := false
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "server:" {
			inServer = true
			continue
		}
		if strings.HasSuffix(line, ":") && !strings.Contains(line, " ") && line != "server:" {
			inServer = false
			continue
		}
		if !inServer || line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value := parseKeyValue(line)
		switch key {
		case "interface":
			cfg.Interface = append(cfg.Interface, value)
		case "port":
			fmt.Sscanf(value, "%d", &cfg.Port)
		case "access-control":
			cfg.AccessControl = append(cfg.AccessControl, value)
		case "do-ip4":
			cfg.DoIPv4 = value == "yes"
		case "do-ip6":
			cfg.DoIPv6 = value == "yes"
		case "do-udp":
			cfg.DoUDP = value == "yes"
		case "do-tcp":
			cfg.DoTCP = value == "yes"
		case "verbosity":
			fmt.Sscanf(value, "%d", &cfg.Verbosity)
		case "num-threads":
			fmt.Sscanf(value, "%d", &cfg.NumThreads)
		case "cache-min-ttl":
			fmt.Sscanf(value, "%d", &cfg.CacheMinTTL)
		case "cache-max-ttl":
			fmt.Sscanf(value, "%d", &cfg.CacheMaxTTL)
		case "prefetch":
			cfg.Prefetch = value == "yes"
		case "hide-identity":
			cfg.HideIdentity = value == "yes"
		case "hide-version":
			cfg.HideVersion = value == "yes"
		}
	}
	return cfg
}

func (cm *ConfigManager) parseForwardZones(content string) []ForwardZone {
	var zones []ForwardZone
	var current *ForwardZone

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "forward-zone:" {
			if current != nil {
				zones = append(zones, *current)
			}
			current = &ForwardZone{}
			continue
		}

		if current == nil {
			continue
		}

		if strings.HasSuffix(line, ":") && !strings.Contains(line, " ") && line != "forward-zone:" {
			zones = append(zones, *current)
			current = nil
			continue
		}

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value := parseKeyValue(line)
		switch key {
		case "name":
			current.Name = strings.Trim(value, "\"")
		case "forward-addr":
			current.Addrs = append(current.Addrs, value)
		case "forward-tls-upstream":
			current.TLSUp = value
		}
	}

	if current != nil {
		zones = append(zones, *current)
	}

	return zones
}

func (cm *ConfigManager) parseIncludes(content string) []string {
	var includes []string
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "include:") {
			path := strings.TrimSpace(strings.TrimPrefix(line, "include:"))
			path = strings.Trim(path, "\"")
			includes = append(includes, path)
		}
	}
	return includes
}

func parseKeyValue(line string) (string, string) {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}
