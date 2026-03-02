package storagecmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/koyeo/nest/config"
	"github.com/koyeo/nest/storage"
	"github.com/koyeo/nest/utils/unit"
	"github.com/spf13/cobra"
)

var (
	flagProvider        string
	flagEndpoint        string
	flagRegion          string
	flagBucket          string
	flagAccessKeyID     string
	flagAccessKeySecret string
)

// Cmd is the root storage command group.
var Cmd = &cobra.Command{
	Use:   "storage",
	Short: "Manage cloud storage configs / 管理云存储配置",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var addCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a storage config / 添加云存储配置",
	Long: `Add a cloud storage configuration.

Interactive mode (guided):
  nest storage add

Non-interactive mode (for scripts / AI):
  nest storage add my-oss \
    --provider oss \
    --endpoint oss-cn-hangzhou.aliyuncs.com \
    --bucket my-deploy-bucket \
    --access-key-id LTAI5t... \
    --access-key-secret xxxxxxxx`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAdd,
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List storage configs / 列出云存储配置",
	RunE:  runList,
}

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a storage config / 删除云存储配置",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

var usageCmd = &cobra.Command{
	Use:   "usage <name>",
	Short: "Show storage space usage / 查看云存储空间使用量",
	Args:  cobra.ExactArgs(1),
	RunE:  runUsage,
}

var cleanCmd = &cobra.Command{
	Use:   "clean <name>",
	Short: "Clean all nest objects from storage / 清空云存储中的 nest 文件",
	Args:  cobra.ExactArgs(1),
	RunE:  runClean,
}

func init() {
	addCmd.Flags().StringVar(&flagProvider, "provider", "", "Storage provider: oss or s3")
	addCmd.Flags().StringVar(&flagEndpoint, "endpoint", "", "Service endpoint (required for OSS)")
	addCmd.Flags().StringVar(&flagRegion, "region", "", "Region (required for S3)")
	addCmd.Flags().StringVar(&flagBucket, "bucket", "", "Bucket name")
	addCmd.Flags().StringVar(&flagAccessKeyID, "access-key-id", "", "Access Key ID")
	addCmd.Flags().StringVar(&flagAccessKeySecret, "access-key-secret", "", "Access Key Secret")
	Cmd.AddCommand(addCmd, listCmd, removeCmd, usageCmd, cleanCmd)
}

func prompt(reader *bufio.Reader, label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("%s: ", label)
	}
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal
	}
	return input
}

func runAdd(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	// Step 1: Storage config name
	var name string
	if len(args) > 0 {
		name = args[0]
	} else if flagProvider != "" && flagBucket != "" {
		// Non-interactive but no name — derive from bucket
		name = flagBucket
	} else {
		fmt.Println("☁️  Add Cloud Storage")
		fmt.Println("──────────────────────────")
		fmt.Println()
		name = prompt(reader, "Config name (e.g. oss-prod, s3-us)", "")
		if name == "" {
			return fmt.Errorf("config name is required")
		}
	}

	// Step 2: Provider
	if flagProvider == "" {
		fmt.Println()
		fmt.Println("Select provider:")
		fmt.Println("  1) oss  — Alibaba Cloud OSS")
		fmt.Println("  2) s3   — AWS S3 (or S3-compatible)")
		fmt.Println()
		choice := prompt(reader, "Provider (1/2 or oss/s3)", "")
		switch choice {
		case "1", "oss":
			flagProvider = "oss"
		case "2", "s3":
			flagProvider = "s3"
		default:
			return fmt.Errorf("invalid provider: '%s' (use 'oss' or 's3')", choice)
		}
	}
	if flagProvider != "oss" && flagProvider != "s3" {
		return fmt.Errorf("provider must be 'oss' or 's3', got '%s'", flagProvider)
	}

	// Step 3: Provider-specific config
	fmt.Println()
	if flagProvider == "oss" {
		if flagEndpoint == "" {
			flagEndpoint = prompt(reader, "OSS Endpoint (e.g. oss-cn-hangzhou.aliyuncs.com)", "")
			if flagEndpoint == "" {
				return fmt.Errorf("endpoint is required for OSS")
			}
		}
	} else {
		if flagRegion == "" {
			flagRegion = prompt(reader, "AWS Region (e.g. us-east-1)", "")
			if flagRegion == "" {
				return fmt.Errorf("region is required for S3")
			}
		}
		if flagEndpoint == "" {
			flagEndpoint = prompt(reader, "Custom endpoint (leave empty for AWS)", "")
		}
	}

	// Step 4: Bucket name
	if flagBucket == "" {
		flagBucket = prompt(reader, "Bucket name", "")
		if flagBucket == "" {
			return fmt.Errorf("bucket name is required")
		}
	}

	// Step 5: Credentials
	fmt.Println()
	if flagAccessKeyID == "" {
		flagAccessKeyID = prompt(reader, "Access Key ID", "")
		if flagAccessKeyID == "" {
			return fmt.Errorf("Access Key ID is required")
		}
	}
	if flagAccessKeySecret == "" {
		flagAccessKeySecret = prompt(reader, "Access Key Secret", "")
		if flagAccessKeySecret == "" {
			return fmt.Errorf("Access Key Secret is required")
		}
	}

	// Save
	cfg := config.Load()
	if err := cfg.AddStorage(name, flagProvider, flagEndpoint, flagRegion, flagBucket, flagAccessKeyID, flagAccessKeySecret); err != nil {
		return err
	}
	if err := config.Save(cfg); err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("✅ Storage '%s' saved (credentials encrypted in ~/.nest/config.json)\n", name)
	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	cfg := config.Load()
	if len(cfg.Storages) == 0 {
		fmt.Println("No storage configs found.")
		fmt.Println()
		fmt.Println("Add one with:")
		fmt.Println("  nest storage add")
		return nil
	}

	fmt.Println("Configured storages:")
	fmt.Println()
	for name, b := range cfg.Storages {
		fmt.Printf("  📦 %s\n", name)
		fmt.Printf("     provider : %s\n", b.Provider)
		fmt.Printf("     bucket   : %s\n", b.BucketName)
		if b.Endpoint != "" {
			fmt.Printf("     endpoint : %s\n", b.Endpoint)
		}
		if b.Region != "" {
			fmt.Printf("     region   : %s\n", b.Region)
		}
		fmt.Printf("     keys     : %s (encrypted)\n", maskSecret(b.AccessKeyID))
		fmt.Println()
	}
	return nil
}

func runRemove(cmd *cobra.Command, args []string) error {
	name := args[0]
	cfg := config.Load()
	if err := cfg.RemoveStorage(name); err != nil {
		return err
	}
	if err := config.Save(cfg); err != nil {
		return err
	}
	fmt.Printf("✅ Storage '%s' removed\n", name)
	return nil
}

func newStorageClient(name string) (storage.ObjectStorage, error) {
	cfg := config.Load()
	cred, err := cfg.DecryptStorage(name)
	if err != nil {
		return nil, fmt.Errorf("storage '%s': %s", name, err)
	}
	return storage.NewFromCredential(cred)
}

func runUsage(cmd *cobra.Command, args []string) error {
	name := args[0]
	store, err := newStorageClient(name)
	if err != nil {
		return err
	}

	ctx := context.Background()
	objects, err := store.ListObjects(ctx, "nest/")
	if err != nil {
		return fmt.Errorf("list objects error: %s", err)
	}

	if len(objects) == 0 {
		fmt.Printf("📦 Storage '%s': empty (no nest/ objects)\n", name)
		return nil
	}

	var totalSize int64
	for _, obj := range objects {
		totalSize += obj.Size
	}

	fmt.Printf("📦 Storage '%s':\n", name)
	fmt.Printf("   Objects : %d\n", len(objects))
	fmt.Printf("   Size    : %s\n", unit.ByteSize(totalSize))
	fmt.Println()

	for _, obj := range objects {
		fmt.Printf("   %s  %s\n", unit.ByteSize(obj.Size), obj.Key)
	}

	return nil
}

func runClean(cmd *cobra.Command, args []string) error {
	name := args[0]
	store, err := newStorageClient(name)
	if err != nil {
		return err
	}

	ctx := context.Background()
	objects, err := store.ListObjects(ctx, "nest/")
	if err != nil {
		return fmt.Errorf("list objects error: %s", err)
	}

	if len(objects) == 0 {
		fmt.Printf("📦 Storage '%s': already empty\n", name)
		return nil
	}

	var totalSize int64
	for _, obj := range objects {
		totalSize += obj.Size
	}

	fmt.Printf("⚠️  About to delete %d objects (%s) from storage '%s'\n", len(objects), unit.ByteSize(totalSize), name)
	fmt.Println()
	for _, obj := range objects {
		fmt.Printf("   🗑  %s  %s\n", unit.ByteSize(obj.Size), obj.Key)
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	confirm := prompt(reader, "Type 'yes' to confirm", "")
	if confirm != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	keys := make([]string, len(objects))
	for i, obj := range objects {
		keys[i] = obj.Key
	}

	if err = store.DeleteObjects(ctx, keys); err != nil {
		return fmt.Errorf("delete objects error: %s", err)
	}

	fmt.Printf("✅ Deleted %d objects (%s)\n", len(objects), unit.ByteSize(totalSize))
	return nil
}

func maskSecret(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
