package unbound

import (
	"fmt"
	"strconv"
	"strings"
)

// Statistics represents parsed unbound statistics
type Statistics struct {
	TotalQueries   int64              `json:"total_queries"`
	CacheHits      int64              `json:"cache_hits"`
	CacheMisses    int64              `json:"cache_misses"`
	CacheHitRatio  float64            `json:"cache_hit_ratio"`
	Uptime         int64              `json:"uptime"`
	MemoryMB       float64            `json:"memory_mb"`
	NumThreads     int                `json:"num_threads"`
	QueryTypes     map[string]int64   `json:"query_types"`
	ResponseCodes  map[string]int64   `json:"response_codes"`
	CurrentQueries int64              `json:"current_queries"`
	TotalRecursion int64              `json:"total_recursion"`
	AvgRecursion   float64            `json:"avg_recursion_ms"`
}

// ParseStatistics converts raw stats map to structured Statistics
func ParseStatistics(raw map[string]string) *Statistics {
	stats := &Statistics{
		QueryTypes:    make(map[string]int64),
		ResponseCodes: make(map[string]int64),
	}

	// Total queries
	stats.TotalQueries = sumThreadStats(raw, "total.num.queries")

	// Cache hits and misses
	stats.CacheHits = sumThreadStats(raw, "total.num.cachehits")
	stats.CacheMisses = sumThreadStats(raw, "total.num.cachemiss")

	// Cache hit ratio
	total := stats.CacheHits + stats.CacheMisses
	if total > 0 {
		stats.CacheHitRatio = float64(stats.CacheHits) / float64(total) * 100
	}

	// Uptime
	if v, ok := raw["time.up"]; ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			stats.Uptime = int64(f)
		}
	}

	// Memory
	stats.MemoryMB = parseMemory(raw)

	// Query types
	for key, value := range raw {
		if strings.HasPrefix(key, "num.query.type.") {
			qtype := strings.TrimPrefix(key, "num.query.type.")
			if v, err := strconv.ParseInt(value, 10, 64); err == nil {
				stats.QueryTypes[qtype] = v
			}
		}
	}

	// Response codes
	for key, value := range raw {
		if strings.HasPrefix(key, "num.answer.rcode.") {
			rcode := strings.TrimPrefix(key, "num.answer.rcode.")
			if v, err := strconv.ParseInt(value, 10, 64); err == nil {
				stats.ResponseCodes[rcode] = v
			}
		}
	}

	// Recursion time
	if v, ok := raw["total.recursion.time.avg"]; ok {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			stats.AvgRecursion = f * 1000 // convert to ms
		}
	}

	return stats
}

func sumThreadStats(raw map[string]string, key string) int64 {
	// Try direct key first
	if v, ok := raw[key]; ok {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}

	// Sum thread-specific keys
	var sum int64
	for i := 0; i < 64; i++ {
		threadKey := fmt.Sprintf("thread%d.%s", i, strings.TrimPrefix(key, "total."))
		if v, ok := raw[threadKey]; ok {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				sum += n
			}
		} else {
			break
		}
	}
	return sum
}

func parseMemory(raw map[string]string) float64 {
	var totalBytes int64
	memKeys := []string{
		"mem.cache.rrset",
		"mem.cache.message",
		"mem.mod.iterator",
		"mem.mod.validator",
	}
	for _, key := range memKeys {
		if v, ok := raw[key]; ok {
			if n, err := strconv.ParseInt(v, 10, 64); err == nil {
				totalBytes += n
			}
		}
	}
	return float64(totalBytes) / 1024 / 1024
}
