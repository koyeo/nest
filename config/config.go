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

// BucketCredential holds cloud storage credentials (access keys are AES-encrypted).
type BucketCredential struct {
	Provider        string `json:"provider"` // "oss" or "s3"
	Endpoint        string `json:"endpoint,omitempty"`
	Region          string `json:"region,omitempty"`
	BucketName      string `json:"bucket"`
	AccessKeyID     string `json:"access_key_id"`     // encrypted
	AccessKeySecret string `json:"access_key_secret"` // encrypted
}

// UserConfig holds user-level preferences stored at ~/.nest/config.json.
type UserConfig struct {
	Lang       string                       `json:"lang"`                  // "zh" or "en"
	EncryptKey string                       `json:"encrypt_key,omitempty"` // base64 AES-256 key
	Buckets    map[string]*BucketCredential `json:"buckets,omitempty"`
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
	return cfg
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

// AddBucket encrypts credentials and stores a bucket config.
func (c *UserConfig) AddBucket(name, provider, endpoint, region, bucketName, accessKeyID, accessKeySecret string) error {
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

	if c.Buckets == nil {
		c.Buckets = make(map[string]*BucketCredential)
	}
	c.Buckets[name] = &BucketCredential{
		Provider:        provider,
		Endpoint:        endpoint,
		Region:          region,
		BucketName:      bucketName,
		AccessKeyID:     encID,
		AccessKeySecret: encSecret,
	}
	return nil
}

// RemoveBucket deletes a bucket config by name.
func (c *UserConfig) RemoveBucket(name string) error {
	if c.Buckets == nil {
		return fmt.Errorf("bucket '%s' not found", name)
	}
	if _, ok := c.Buckets[name]; !ok {
		return fmt.Errorf("bucket '%s' not found", name)
	}
	delete(c.Buckets, name)
	return nil
}

// DecryptBucket returns a copy of BucketCredential with plaintext access keys.
func (c *UserConfig) DecryptBucket(name string) (*BucketCredential, error) {
	if c.Buckets == nil {
		return nil, fmt.Errorf("bucket '%s' not found", name)
	}
	b, ok := c.Buckets[name]
	if !ok {
		return nil, fmt.Errorf("bucket '%s' not found", name)
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

	return &BucketCredential{
		Provider:        b.Provider,
		Endpoint:        b.Endpoint,
		Region:          b.Region,
		BucketName:      b.BucketName,
		AccessKeyID:     id,
		AccessKeySecret: secret,
	}, nil
}
