// Package goscribe provides a public API for generating and updating
// documentation from source code using AI providers.
//
// This package is the stable programmatic interface for GoScribe. It wraps
// internal functionality without exposing internal types, and is designed
// to remain backward-compatible once the module reaches v1.0.0.
//
// Until v1.0.0, this API should be considered experimental.
package goscribe

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/house/goscribe/internal/ai"
	"github.com/house/goscribe/internal/config"
	"github.com/house/goscribe/internal/docs"
	"github.com/house/goscribe/internal/git"
	"github.com/spf13/viper"
)

// Provider generates text from a prompt.
//
// This mirrors the internal provider contract without exposing internal/ai.
type Provider interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

// Client is the public entry point for generating and updating documentation.
type Client struct {
	provider     Provider
	providerName string
	model        string
	outputDir    string
	timeout      time.Duration
	retries      int
	retryBackoff time.Duration
	verbose      bool
	workers      int
	cacheDir     string
	profile      string
}

// NewClient creates a Client using the supplied options.
//
// If no provider is injected, the client resolves the configured provider
// using the same provider configuration used by the CLI.
func NewClient(opts ...ClientOption) (*Client, error) {
	var co clientOptions
	for _, opt := range opts {
		opt(&co)
	}

	// Apply defaults
	if co.outputDir == "" {
		co.outputDir = "docs"
	}
	if co.timeout == 0 {
		co.timeout = 5 * time.Minute
	}
	if co.retries == 0 {
		co.retries = 3
	}
	if co.retryBackoff == 0 {
		co.retryBackoff = 2 * time.Second
	}

	c := &Client{
		provider:     co.provider,
		providerName: co.providerName,
		model:        co.model,
		outputDir:    co.outputDir,
		timeout:      co.timeout,
		retries:      co.retries,
		retryBackoff: co.retryBackoff,
		verbose:      co.verbose,
		workers:      co.workers,
		cacheDir:     co.cacheDir,
		profile:      co.profile,
	}

	// If a provider was injected, use it directly.
	if c.provider != nil {
		return c, nil
	}

	// Load configuration to resolve provider.
	if err := loadConfig(co.configFile); err != nil {
		return nil, &Error{
			Op:   "load_config",
			Kind: ErrGenerationFailed,
			Err:  err,
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, &Error{
			Op:   "load_config",
			Kind: ErrGenerationFailed,
			Err:  err,
		}
	}

	// Override config with option values
	if c.providerName != "" {
		cfg.Provider = c.providerName
	}
	if c.model != "" {
		cfg.Model = c.model
	}
	if co.timeout != 0 {
		cfg.Timeout = c.timeout
	}
	if co.retries != 0 {
		cfg.Retries = c.retries
	}
	if co.retryBackoff != 0 {
		cfg.RetryBackoff = c.retryBackoff
	}
	if co.verbose {
		cfg.Verbose = c.verbose
	}
	if co.profile == "" && cfg.Profile != "" {
		c.profile = cfg.Profile
	}

	internalProvider, err := ai.NewProvider(cfg)
	if err != nil {
		kind := ErrProviderNotConfigured
		if c.providerName != "" {
			kind = ErrUnsupportedProvider
		}
		return nil, &Error{
			Op:   "resolve_provider",
			Kind: kind,
			Err:  err,
		}
	}

	c.provider = internalProvider
	if cfg.Provider != "" {
		c.providerName = cfg.Provider
	}
	if cfg.Model != "" {
		c.model = cfg.Model
	}

	return c, nil
}

// loadConfig initializes viper with the same semantics as the CLI.
func loadConfig(configFile string) error {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home dir: %w", err)
		}
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".goscribe")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("GOSCRIBE")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("read config: %w", err)
		}
	}

	return nil
}

// resolveOutputDir returns the effective output directory for an operation.
func (c *Client) resolveOutputDir(override string) string {
	if override != "" {
		return override
	}
	return c.outputDir
}

// saveStateIfGit saves the current commit hash to .goscribe-state when
// sourcePath is a git repository.
func (c *Client) saveStateIfGit(sourcePath string) (string, bool, error) {
	repo, err := git.OpenRepo(sourcePath)
	if err != nil {
		return "", false, nil // not a git repo, not an error
	}

	commit, err := repo.GetCurrentCommit()
	if err != nil {
		return "", false, err
	}

	if err := docs.SaveCommitState(sourcePath, commit); err != nil {
		return commit, false, err
	}

	return commit, true, nil
}
