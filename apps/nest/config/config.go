package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	nestDir    = ".nest"
	configFile = "config.json"
)

// StorageCredential holds cloud storage credentials (access keys are AES-encrypted).
type StorageCredential struct {
	Provider        string `json:"provider"` // "oss" or "s3"
	Endpoint        string `json:"endpoint,omitempty"`
	Region          string `json:"region,omitempty"`
	BucketName      string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`     // encrypted
	AccessKeySecret string `json:"access_key_secret"` // encrypted
}

// UserConfig holds user-level preferences stored at ~/.nest/config.json.
type UserConfig struct {
	Lang       string                        `json:"lang"`                  // "zh" or "en"
	EncryptKey string                        `json:"encrypt_key,omitempty"` // base64 AES-256 key
	Storages   map[string]*StorageCredential `json:"storages,omitempty"`
}

func defaultConfig() *UserConfig {
	return &UserConfig{Lang: "zh"}
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home dir error: %s", err)
	}
	return filepath.Join(home, nestDir, configFile), nil
}

// Load reads ~/.nest/config.json. Falls back to defaults on any error.
func Load() *UserConfig {
	p, err := configPath()
	if err != nil {
		return defaultConfig()
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return defaultConfig()
	}
	cfg := &UserConfig{}
	if err = json.Unmarshal(data, cfg); err != nil {
		return defaultConfig()
	}
	if cfg.Lang != "zh" && cfg.Lang != "en" {
		cfg.Lang = "zh"
	}
	// Backward compatibility: migrate "buckets" → "storages"
	if cfg.Storages == nil {
		cfg.Storages = migrateOldBuckets(data)
	}
	return cfg
}

// migrateOldBuckets reads the legacy "buckets" key from raw JSON if present.
func migrateOldBuckets(data []byte) map[string]*StorageCredential {
	var legacy struct {
		Buckets map[string]*StorageCredential `json:"buckets"`
	}
	if err := json.Unmarshal(data, &legacy); err == nil && len(legacy.Buckets) > 0 {
		return legacy.Buckets
	}
	return nil
}

// Save writes the config to ~/.nest/config.json.
func Save(cfg *UserConfig) error {
	p, err := configPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(p)
	if err = os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config dir error: %s", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config error: %s", err)
	}
	return os.WriteFile(p, data, 0644)
}

// EnsureEncryptKey generates an encryption key if one doesn't exist yet.
func (c *UserConfig) EnsureEncryptKey() error {
	if c.EncryptKey != "" {
		return nil
	}
	key, err := GenerateEncryptKey()
	if err != nil {
		return err
	}
	c.EncryptKey = key
	return nil
}

// AddStorage encrypts credentials and stores a storage config.
func (c *UserConfig) AddStorage(name, provider, endpoint, region, bucketName, accessKeyID, accessKeySecret string) error {
	if err := c.EnsureEncryptKey(); err != nil {
		return err
	}

	encID, err := Encrypt(c.EncryptKey, accessKeyID)
	if err != nil {
		return fmt.Errorf("encrypt access_key_id error: %s", err)
	}
	encSecret, err := Encrypt(c.EncryptKey, accessKeySecret)
	if err != nil {
		return fmt.Errorf("encrypt access_key_secret error: %s", err)
	}

	if c.Storages == nil {
		c.Storages = make(map[string]*StorageCredential)
	}
	c.Storages[name] = &StorageCredential{
		Provider:        provider,
		Endpoint:        endpoint,
		Region:          region,
		BucketName:      bucketName,
		AccessKeyID:     encID,
		AccessKeySecret: encSecret,
	}
	return nil
}

// RemoveStorage deletes a storage config by name.
func (c *UserConfig) RemoveStorage(name string) error {
	if c.Storages == nil {
		return fmt.Errorf("storage '%s' not found", name)
	}
	if _, ok := c.Storages[name]; !ok {
		return fmt.Errorf("storage '%s' not found", name)
	}
	delete(c.Storages, name)
	return nil
}

// DecryptStorage returns a copy of StorageCredential with plaintext access keys.
func (c *UserConfig) DecryptStorage(name string) (*StorageCredential, error) {
	if c.Storages == nil {
		return nil, fmt.Errorf("storage '%s' not found", name)
	}
	b, ok := c.Storages[name]
	if !ok {
		return nil, fmt.Errorf("storage '%s' not found", name)
	}
	if c.EncryptKey == "" {
		return nil, fmt.Errorf("encrypt_key not set")
	}

	id, err := Decrypt(c.EncryptKey, b.AccessKeyID)
	if err != nil {
		return nil, fmt.Errorf("decrypt access_key_id error: %s", err)
	}
	secret, err := Decrypt(c.EncryptKey, b.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("decrypt access_key_secret error: %s", err)
	}

	return &StorageCredential{
		Provider:        b.Provider,
		Endpoint:        b.Endpoint,
		Region:          b.Region,
		BucketName:      b.BucketName,
		AccessKeyID:     id,
		AccessKeySecret: secret,
	}, nil
}
