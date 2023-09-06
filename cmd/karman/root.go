package main

import (
	"encoding"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:               "karman",
	Short:             "Karman - The Karaoke Manager",
	Long:              `The Karaoke Manager helps you organize your UltraStar Karaoke songs.`,
	SilenceUsage:      true,
	DisableAutoGenTag: true,
	Args:              cobra.NoArgs,
	Version:           version,
	PersistentPreRunE: loadConfig,
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "", "Custom config file")

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
	config struct {
		Log struct {
			Level  slog.Level `mapstructure:"level"`
			Format string     `mapstructure:"format"`
		} `mapstructure:"log"`
		DBConnection string `mapstructure:"db-url"`
		API          struct {
			Address string `mapstructure:"address"`
		} `mapstructure:"api"`
		TaskServer struct {
			Workers int `mapstructure:"workers"`
		} `mapstructure:"task-server"`
		Uploads struct {
			Dir string `mapstructure:"dir"`
		} `mapstructure:"uploads"`
		Media struct {
			Dir string `mapstructure:"dir"`
		} `mapstructure:"media"`
	}
	logger *slog.Logger
)

func loadConfig(cmd *cobra.Command, _ []string) error {
	if cmd == versionCmd {
		// do not load config for version command
		return nil
	}
	// we don't allow HCL-style configs
	viper.SupportedExts = []string{"json", "toml", "yaml", "yml", "env", "ini"}
	viper.AllowEmptyEnv(true)
	viper.SetEnvPrefix("karman")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()
	configFile := cmd.Flag("config").Value.String()
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/karman/")
	}

	// TODO: Hot Reloading
	if err := viper.ReadInConfig(); err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) {
		return err
	}

	var meta mapstructure.Metadata
	if err := viper.Unmarshal(&config, func(config *mapstructure.DecoderConfig) {
		config.WeaklyTypedInput = true
		config.Metadata = &meta
		config.DecodeHook = mapstructure.DecodeHookFunc(func(from reflect.Value, to reflect.Value) (interface{}, error) {
			if to.CanAddr() {
				to = to.Addr()
			}
			// If the destination implements the unmarshaling interface
			u, ok := to.Interface().(encoding.TextUnmarshaler)
			if !ok {
				return from.Interface(), nil
			}
			// If it is nil and a pointer, create and assign the target value first
			if to.IsNil() && to.Type().Kind() == reflect.Ptr {
				to.Set(reflect.New(to.Type().Elem()))
				u = to.Interface().(encoding.TextUnmarshaler)
			}
			var text []byte
			switch v := from.Interface().(type) {
			case string:
				text = []byte(v)
			case []byte:
				text = v
			default:
				return v, nil
			}

			if err := u.UnmarshalText(text); err != nil {
				return to.Interface(), err
			}
			return to.Interface(), nil
		})
	}); err != nil {
		return fmt.Errorf("unable to decode config file: %w", err)
	}
	if len(meta.Unused) == 1 {
		return fmt.Errorf("invalid key in config file: %s", meta.Unused[0])
	} else if len(meta.Unused) > 0 {
		return fmt.Errorf("invalid keys in config file: %s", strings.Join(meta.Unused, ", "))
	}

	config.Log.Format = strings.ToLower(config.Log.Format)
	if config.Log.Format == "text" {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: config.Log.Level}))
	} else if config.Log.Format == "json" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: config.Log.Level}))
	} else if config.Log.Format == "color" {
		logger = slog.New(tint.NewHandler(colorable.NewColorableStdout(), &tint.Options{Level: config.Log.Level}))
	} else {
		return fmt.Errorf("invalid log format: %s", viper.GetString("log.format"))
	}

	if viper.ConfigFileUsed() != "" {
		logger.Info("Loaded configuration file", "file", viper.ConfigFileUsed())
	} else {
		logger.Info("No configuration file found")
	}
	return nil
}
