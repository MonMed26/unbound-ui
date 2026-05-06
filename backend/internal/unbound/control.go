package unbound

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Control wraps unbound-control commands
type Control struct {
	BinaryPath string
}

func NewControl(binaryPath string) *Control {
	if binaryPath == "" {
		binaryPath = "unbound-control"
	}
	return &Control{BinaryPath: binaryPath}
}

// Status returns the status of unbound
func (c *Control) Status() (string, error) {
	return c.exec("status")
}

// Stats returns unbound statistics without resetting counters
func (c *Control) Stats() (map[string]string, error) {
	output, err := c.exec("stats_noreset")
	if err != nil {
		return nil, err
	}
	return parseStats(output), nil
}

// Reload reloads the unbound configuration
func (c *Control) Reload() error {
	_, err := c.exec("reload")
	return err
}

// FlushCache flushes the DNS cache
func (c *Control) FlushCache() error {
	_, err := c.exec("flush_zone", ".")
	return err
}

// LocalZoneAdd adds a local zone
func (c *Control) LocalZoneAdd(zone, zoneType string) error {
	_, err := c.exec("local_zone", zone, zoneType)
	return err
}

// LocalZoneRemove removes a local zone
func (c *Control) LocalZoneRemove(zone string) error {
	_, err := c.exec("local_zone_remove", zone)
	return err
}

// LocalDataAdd adds local data
func (c *Control) LocalDataAdd(data string) error {
	_, err := c.exec("local_data", data)
	return err
}

// LocalDataRemove removes local data
func (c *Control) LocalDataRemove(name string) error {
	_, err := c.exec("local_data_remove", name)
	return err
}

// DumpCache dumps the cache contents
func (c *Control) DumpCache() (string, error) {
	return c.exec("dump_cache")
}

// ListLocalZones lists all local zones
func (c *Control) ListLocalZones() ([]LocalZone, error) {
	output, err := c.exec("list_local_zones")
	if err != nil {
		return nil, err
	}
	return parseLocalZones(output), nil
}

// ListLocalData lists all local data
func (c *Control) ListLocalData() ([]string, error) {
	output, err := c.exec("list_local_data")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var result []string
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			result = append(result, line)
		}
	}
	return result, nil
}

func (c *Control) exec(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, c.BinaryPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("unbound-control %s failed: %w (output: %s)", args[0], err, string(output))
	}
	return string(output), nil
}

func parseStats(output string) map[string]string {
	stats := make(map[string]string)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			stats[key] = value
		}
	}
	return stats
}

type LocalZone struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func parseLocalZones(output string) []LocalZone {
	var zones []LocalZone
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			zones = append(zones, LocalZone{
				Name: parts[0],
				Type: parts[1],
			})
		}
	}
	return zones
}
