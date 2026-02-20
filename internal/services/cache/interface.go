package cache

// CacheInterface defines the contract for prediction caching.
// Using an interface allows handlers to be tested with mock implementations.
type CacheInterface interface {
	Get(ticker string) (map[string]interface{}, bool)
	Set(ticker string, data map[string]interface{}) error
	Delete(ticker string) error
	GetCachedTickers() ([]string, error)
	GetForTicker(ticker string) (map[string]interface{}, error)
}
