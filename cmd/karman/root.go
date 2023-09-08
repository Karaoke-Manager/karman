package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Karaoke-Manager/karman/cmd/karman/internal"
)

// rootCmd represents the main "karman" command.
// The command cannot be executed by itself.
var rootCmd = &cobra.Command{
	Use:               "karman",
	Short:             "Karman - The Karaoke Manager",
	Long:              `The Karaoke Manager helps you organize your UltraStar Karaoke songs.`,
	SilenceUsage:      true,
	DisableAutoGenTag: true,
	Args:              cobra.NoArgs,
	Version:           version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd == versionCmd {
			// do not load config for version command
			return nil
		}
		if err := loadConfig(); err != nil {
			return err
		}
		if err := setupLogger(); err != nil {
			return err
		}
		if viper.ConfigFileUsed() != "" {
			mainLogger.Info(fmt.Sprintf("Using configuration file %s.", viper.ConfigFileUsed()))
		} else {
			mainLogger.Info("No configuration file found.")
		}
		if config.Debug {
			mainLogger.Warn("Debug mode is enabled.")
		}
		return nil
	},
}

// init sets up common flags for all other commands.
func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Custom config file")

	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug mode.")
	_ = viper.BindPFlag("debug", rootCmd.Flag("debug"))
	viper.SetDefault("debug", false)

	rootCmd.PersistentFlags().String("log-level", slog.LevelInfo.String(), "The logging verbosity. Can be set to DEBUG, INFO, WARN, ERROR or an integer where lower numbers mean more logging.")
	_ = viper.BindPFlag("log.level", rootCmd.Flag("log-level"))
	viper.SetDefault("log.level", 0)

	rootCmd.PersistentFlags().String("log-format", "text", `Format used for logging. Allowed values are "text" or "json".`)
	_ = viper.BindPFlag("log.format", rootCmd.Flag("log-format"))
	viper.SetDefault("log.format", "text")

	rootCmd.PersistentFlags().String("db-url", "", "PostgreSQL Connection String")
	_ = viper.BindPFlag("db-url", rootCmd.Flag("db-url"))
	viper.SetDefault("db-url", "")
}

var (
	// configFile is a path to a config file passed as CLI argument.
	configFile string
	// config is the parsed configuration from files, environment and flags.
	config struct {
		Debug bool `mapstructure:"debug"`
		Log   struct {
			Level  slog.Level `mapstructure:"level"`
			Format string     `mapstructure:"format"`
		} `mapstructure:"log"`
		DBConnection string `mapstructure:"db-url"`
		API          struct {
			Address string `mapstructure:"address"`
		} `mapstructure:"api"`
		TaskRunner struct {
			Workers int `mapstructure:"workers"`
		} `mapstructure:"task-server"`
		Uploads struct {
			Dir string `mapstructure:"dir"`
		} `mapstructure:"uploads"`
		Media struct {
			Dir string `mapstructure:"dir"`
		} `mapstructure:"media"`
	}
)

// logger is the global logger.
var (
	logger     *slog.Logger
	mainLogger *slog.Logger
)

// loadConfig parses the configuration file and merges it with configuration data
// from the environment and CLI flags.
func loadConfig() error {
	// we don't allow HCL-style configs
	viper.SupportedExts = []string{"json", "toml", "yaml", "yml", "env", "ini"}
	viper.AllowEmptyEnv(true)
	viper.SetEnvPrefix("karman")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/karman/")
	}

	if err := viper.ReadInConfig(); err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) {
		return err
	}

	var meta mapstructure.Metadata
	if err := viper.Unmarshal(&config, func(config *mapstructure.DecoderConfig) {
		config.WeaklyTypedInput = true
		config.Metadata = &meta
		config.DecodeHook = internal.TextUnmarshalerDecodeHook
	}); err != nil {
		return fmt.Errorf("unable to decode config file: %w", err)
	}
	if len(meta.Unused) == 1 {
		return fmt.Errorf("invalid key in config file: %s", meta.Unused[0])
	} else if len(meta.Unused) > 0 {
		return fmt.Errorf("invalid keys in config file: %s", strings.Join(meta.Unused, ", "))
	}
	return nil
}

// setupLogger sets up the global logger using the app's configuration.
func setupLogger() error {
	config.Log.Format = strings.ToLower(config.Log.Format)
	if config.Log.Format == "text" {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: config.Debug,
			Level:     config.Log.Level,
		}))
	} else if config.Log.Format == "json" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: config.Debug,
			Level:     config.Log.Level,
		}))
	} else if config.Log.Format == "color" {
		logger = slog.New(tint.NewHandler(colorable.NewColorableStdout(), &tint.Options{
			AddSource: config.Debug,
			Level:     config.Log.Level,
		}))
	} else {
		return fmt.Errorf("invalid log format: %s", viper.GetString("log.format"))
	}
	mainLogger = logger.With("log", "main")
	slog.SetDefault(mainLogger)
	return nil
}
