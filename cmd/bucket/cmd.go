package bucket

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/koyeo/nest/config"
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

// Cmd is the root bucket command group.
var Cmd = &cobra.Command{
	Use:   "bucket",
	Short: "Manage cloud storage buckets / 管理云存储配置",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

var addCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a bucket config / 添加云存储配置",
	Long: `Add a cloud storage bucket configuration.

Interactive mode (guided):
  nest bucket add

Non-interactive mode (for scripts / AI):
  nest bucket add my-oss \
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
	Short: "List bucket configs / 列出云存储配置",
	RunE:  runList,
}

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a bucket config / 删除云存储配置",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

func init() {
	addCmd.Flags().StringVar(&flagProvider, "provider", "", "Storage provider: oss or s3")
	addCmd.Flags().StringVar(&flagEndpoint, "endpoint", "", "Service endpoint (required for OSS)")
	addCmd.Flags().StringVar(&flagRegion, "region", "", "Region (required for S3)")
	addCmd.Flags().StringVar(&flagBucket, "bucket", "", "Bucket name")
	addCmd.Flags().StringVar(&flagAccessKeyID, "access-key-id", "", "Access Key ID")
	addCmd.Flags().StringVar(&flagAccessKeySecret, "access-key-secret", "", "Access Key Secret")
	Cmd.AddCommand(addCmd, listCmd, removeCmd)
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

	// Step 1: Bucket config name
	var name string
	if len(args) > 0 {
		name = args[0]
	} else if flagProvider != "" && flagBucket != "" {
		// Non-interactive but no name — derive from bucket
		name = flagBucket
	} else {
		fmt.Println("🪣 Add Cloud Storage Bucket")
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
	if err := cfg.AddBucket(name, flagProvider, flagEndpoint, flagRegion, flagBucket, flagAccessKeyID, flagAccessKeySecret); err != nil {
		return err
	}
	if err := config.Save(cfg); err != nil {
		return err
	}

	fmt.Println()
	fmt.Printf("✅ Bucket '%s' saved (credentials encrypted in ~/.nest/config.json)\n", name)
	return nil
}

func runList(cmd *cobra.Command, args []string) error {
	cfg := config.Load()
	if len(cfg.Buckets) == 0 {
		fmt.Println("No buckets configured.")
		fmt.Println()
		fmt.Println("Add one with:")
		fmt.Println("  nest bucket add")
		return nil
	}

	fmt.Println("Configured buckets:")
	fmt.Println()
	for name, b := range cfg.Buckets {
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
	if err := cfg.RemoveBucket(name); err != nil {
		return err
	}
	if err := config.Save(cfg); err != nil {
		return err
	}
	fmt.Printf("✅ Bucket '%s' removed\n", name)
	return nil
}

func maskSecret(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}
