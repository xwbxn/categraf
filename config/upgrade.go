package config

type UpgradeConfig struct {
	Enable        bool   `boml:"enable"`
	Url           string `toml:"url"`
	Interval      int64  `toml:"interval"`
	BasicAuthUser string `toml:"basic_auth_user"`
	BasicAuthPass string `toml:"basic_auth_pass"`
}
