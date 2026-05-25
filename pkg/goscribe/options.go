package goscribe

import "time"

// ClientOption configures a Client using the functional options pattern.
type ClientOption func(*clientOptions)

// clientOptions holds the configuration values set by ClientOption functions.
type clientOptions struct {
	provider     Provider
	providerName string
	model        string
	outputDir    string
	configFile   string
	timeout      time.Duration
	retries      int
	retryBackoff time.Duration
	verbose      bool
	workers      int
	cacheDir     string
	profile      string
	template     string
}

// WithProvider injects a custom Provider implementation.
func WithProvider(provider Provider) ClientOption {
	return func(o *clientOptions) {
		o.provider = provider
	}
}

// WithConfiguredProvider selects a provider by name from the configuration.
func WithConfiguredProvider(name string) ClientOption {
	return func(o *clientOptions) {
		o.providerName = name
	}
}

// WithModel sets the AI model to use for generation.
func WithModel(model string) ClientOption {
	return func(o *clientOptions) {
		o.model = model
	}
}

// WithOutputDir sets the output directory for generated documentation.
func WithOutputDir(outputDir string) ClientOption {
	return func(o *clientOptions) {
		o.outputDir = outputDir
	}
}

// WithConfigFile sets an explicit configuration file path.
func WithConfigFile(path string) ClientOption {
	return func(o *clientOptions) {
		o.configFile = path
	}
}

// WithTimeout sets the timeout for AI provider calls.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.timeout = timeout
	}
}

// WithRetry sets the retry count and initial backoff duration.
func WithRetry(retries int, backoff time.Duration) ClientOption {
	return func(o *clientOptions) {
		o.retries = retries
		o.retryBackoff = backoff
	}
}

// WithVerbose enables or disables verbose logging output.
func WithVerbose(verbose bool) ClientOption {
	return func(o *clientOptions) {
		o.verbose = verbose
	}
}

// WithWorkers sets the number of concurrent workers for generation/update.
func WithWorkers(n int) ClientOption {
	return func(o *clientOptions) {
		o.workers = n
	}
}

// WithCacheDir sets the directory for content-addressed caching.
func WithCacheDir(dir string) ClientOption {
	return func(o *clientOptions) {
		o.cacheDir = dir
	}
}

// WithProfile sets the documentation profile for generation and updates.
func WithProfile(name string) ClientOption {
	return func(o *clientOptions) {
		o.profile = name
	}
}

// WithTemplate sets the documentation template for generation and updates.
func WithTemplate(name string) ClientOption {
	return func(o *clientOptions) {
		o.template = name
	}
}
