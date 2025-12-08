package openholidays

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	baseURL      = "https://openholidaysapi.org"
	cacheTTL     = 24 * time.Hour
	httpTimeout  = 10 * time.Second
)

// PublicHoliday represents a holiday from the OpenHolidays API
type PublicHoliday struct {
	ID         string          `json:"id"`
	StartDate  string          `json:"startDate"`
	EndDate    string          `json:"endDate"`
	Name       []LocalizedName `json:"name"`
	Nationwide bool            `json:"nationwide"`
	Type       string          `json:"type"`
}

// LocalizedName represents a localized name with language code
type LocalizedName struct {
	Language string `json:"language"`
	Text     string `json:"text"`
}

// cacheEntry holds cached data with expiration
type cacheEntry struct {
	data      []PublicHoliday
	expiresAt time.Time
}

// Client is an HTTP client for the OpenHolidays API
type Client struct {
	httpClient *http.Client
	logger     *zap.Logger
	cache      map[string]cacheEntry
	cacheMu    sync.RWMutex
}

// NewClient creates a new OpenHolidays API client
func NewClient(logger *zap.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: httpTimeout,
		},
		logger: logger,
		cache:  make(map[string]cacheEntry),
	}
}

// GetPublicHolidays fetches public holidays for a country and year
func (c *Client) GetPublicHolidays(ctx context.Context, countryCode string, year int) ([]PublicHoliday, bool, error) {
	cacheKey := fmt.Sprintf("%s-%d", countryCode, year)

	// Check cache first
	c.cacheMu.RLock()
	if entry, ok := c.cache[cacheKey]; ok && time.Now().Before(entry.expiresAt) {
		c.cacheMu.RUnlock()
		c.logger.Debug("returning cached holidays", zap.String("country", countryCode), zap.Int("year", year))
		return entry.data, true, nil
	}
	c.cacheMu.RUnlock()

	// Fetch from API
	url := fmt.Sprintf("%s/PublicHolidays?countryIsoCode=%s&validFrom=%d-01-01&validTo=%d-12-31&languageIsoCode=EN",
		baseURL, countryCode, year, year)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return c.getCachedOrError(cacheKey, err)
	}

	req.Header.Set("Accept", "application/json")

	c.logger.Debug("fetching holidays from API", zap.String("url", url))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Warn("failed to fetch holidays from API", zap.Error(err))
		return c.getCachedOrError(cacheKey, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("API returned status %d", resp.StatusCode)
		c.logger.Warn("holidays API error", zap.Int("status", resp.StatusCode))
		return c.getCachedOrError(cacheKey, err)
	}

	var holidays []PublicHoliday
	if err := json.NewDecoder(resp.Body).Decode(&holidays); err != nil {
		c.logger.Warn("failed to decode holidays response", zap.Error(err))
		return c.getCachedOrError(cacheKey, err)
	}

	// Update cache
	c.cacheMu.Lock()
	c.cache[cacheKey] = cacheEntry{
		data:      holidays,
		expiresAt: time.Now().Add(cacheTTL),
	}
	c.cacheMu.Unlock()

	c.logger.Info("fetched holidays from API", zap.String("country", countryCode), zap.Int("year", year), zap.Int("count", len(holidays)))

	return holidays, false, nil
}

// getCachedOrError returns cached data if available, otherwise returns the error
func (c *Client) getCachedOrError(cacheKey string, originalErr error) ([]PublicHoliday, bool, error) {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()

	// Return expired cache if available (graceful degradation)
	if entry, ok := c.cache[cacheKey]; ok {
		c.logger.Info("returning expired cached holidays due to API error", zap.String("key", cacheKey))
		return entry.data, true, nil
	}

	return nil, false, originalErr
}

// GetName returns the English name from a list of localized names
func GetName(names []LocalizedName) string {
	// Prefer English
	for _, n := range names {
		if n.Language == "EN" {
			return n.Text
		}
	}
	// Fall back to first available
	if len(names) > 0 {
		return names[0].Text
	}
	return ""
}
