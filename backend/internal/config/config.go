package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

// Package config provides configuration loading and saving functionality
// for the Factorio Server Manager application. It supports INI-based config
// files and maps config sections into structured Go types.

// FactorioConfig holds configuration values from the [factorio] section.
type FactorioConfig struct {
	AutoStart       bool          `ini:"auto_start"`                          // Whether the server should auto-start
	Bind            string        `ini:"bind"`                                // Network bind address
	ConfigDir       string        `ini:"config" default:"./config"`           // Path to config directory
	Downloads       string        `ini:"downloads"`                           // Path to download directory
	Files           FactorioFiles `ini:"-"`                                   // Derived file paths (not persisted)
	LogsDir         string        `ini:"logs" default:"./logs"`               // Path to logs directory
	ModsDir         string        `ini:"mods" default:"./mods"`               // Path to mods directory
	SavesDir        string        `ini:"saves" default:"./saves"`             // Path to saves directory
	Save            string        `ini:"save"`                                // Name of the active save file
	SelectedBranch  string        `ini:"branch"`                              // Selected branch (e.g. stable/experimental)
	SelectedVersion string        `ini:"version"`                             // Selected version string
	ServerVersions  string        `ini:"server_versions" default:"./servers"` // Path to downloaded server versions
	Token           string        `ini:"token"`                               // Your factorio.com account API token (https://factorio.com/profile)
	Username        string        `ini:"username"`                            // Your factorio.com account username
}

// FactorioFiles holds derived paths to individual Factorio config files.
type FactorioFiles struct {
	AdminList      string // Path to server-adminlist.json
	BanList        string // Path to server-banlist.json
	ServerId       string // Path to server-id.json
	ServerSettings string // Path to server-settings.json
	WhiteList      string // Path to server-whitelist.json
}

// FSMConfig contains all configuration used by the application.
type FSMConfig struct {
	Admins   map[string]string // Admin usernames and password hashes
	Factorio FactorioConfig    // Factorio configuration
	Path     string            // Path to the loaded config file
	RCon     RConConfig        // RCON configuration
	Server   ServerConfig      // HTTP server configuration
	file     *ini.File         // Internal INI file reference
}

// RConConfig holds configuration for the RCON remote console.
type RConConfig struct {
	Bind     string `ini:"bind" default:"127.0.0.1:27015"` // Bind address for RCON
	Enabled  bool   `ini:"-"`                              // Whether RCON is enabled
	Password string `ini:"password"`                       // RCON password
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Listen string `ini:"listen"` // Listen address for the HTTP server
}

// Load reads the config from disk and parses it into structured config.
func Load(configPath *string) (error, *FSMConfig) {
	resolvedPath := findConfigPath(*configPath)
	if resolvedPath == "" {
		log.Fatal("No config file found.")
	}

	cfg, err := ini.Load(resolvedPath)
	if err != nil {
		return err, nil
	}

	var factorioConfig FactorioConfig
	if err := cfg.Section("factorio").MapTo(&factorioConfig); err != nil {
		return fmt.Errorf("failed to load [factorio]: %w", err), nil
	}

	// Environment variable overrides (useful for Unraid / Docker integration)
	if envUsername := os.Getenv("FACTORIO_USERNAME"); envUsername != "" {
		factorioConfig.Username = envUsername
	}
	if envToken := os.Getenv("FACTORIO_TOKEN"); envToken != "" {
		factorioConfig.Token = envToken
	}

	factorioConfig.Files = FactorioFiles{
		AdminList:      fmt.Sprintf("%s/server-adminlist.json", factorioConfig.ConfigDir),
		BanList:        fmt.Sprintf("%s/server-banlist.json", factorioConfig.ConfigDir),
		ServerId:       fmt.Sprintf("%s/server-id.json", factorioConfig.ConfigDir),
		ServerSettings: fmt.Sprintf("%s/server-settings.json", factorioConfig.ConfigDir),
		WhiteList:      fmt.Sprintf("%s/server-whitelist.json", factorioConfig.ConfigDir),
	}

	var rconConfig RConConfig
	if cfg.HasSection("rcon") {
		if err := cfg.Section("rcon").MapTo(&rconConfig); err != nil {
			return fmt.Errorf("failed to load [rcon]: %w", err), nil
		}
		rconConfig.Enabled = true
	}

	var serverConfig ServerConfig
	if err := cfg.Section("server").MapTo(&serverConfig); err != nil {
		serverConfig.Listen = ":8888" // Unraid Optimization: Default to 8888
	}

	admins := map[string]string{}
	if cfg.HasSection("admins") {
		for _, key := range cfg.Section("admins").Keys() {
			admins[key.Name()] = key.Value()
		}
	}

	fsmConfig := FSMConfig{
		Admins:   admins,
		Factorio: factorioConfig,
		Path:     resolvedPath,
		RCon:     rconConfig,
		Server:   serverConfig,
		file:     cfg,
	}

	// Auto-create required Factorio directories if they do not exist
	dirs := []string{
		fsmConfig.Factorio.ConfigDir,
		fsmConfig.Factorio.Downloads,
		fsmConfig.Factorio.LogsDir,
		fsmConfig.Factorio.ModsDir,
		fsmConfig.Factorio.SavesDir,
		fsmConfig.Factorio.ServerVersions,
		filepath.Dir(fsmConfig.Factorio.Files.AdminList),
	}
	for _, d := range dirs {
		if d != "" {
			os.MkdirAll(d, 0755)
		}
	}

	return nil, &fsmConfig
}

// SaveToFile writes the current config back to the original config path.
func (cfg *FSMConfig) SaveToFile() error {
	if err := cfg.file.Section("factorio").ReflectFrom(&cfg.Factorio); err != nil {
		return fmt.Errorf("failed to write [factorio] config: %w", err)
	}
	if err := cfg.file.Section("rcon").ReflectFrom(&cfg.RCon); err != nil {
		return fmt.Errorf("failed to write [rcon] config: %w", err)
	}
	if err := cfg.file.Section("server").ReflectFrom(&cfg.Server); err != nil {
		return fmt.Errorf("failed to write [server] config: %w", err)
	}
	adminSection := cfg.file.Section("admins")
	for _, key := range adminSection.KeyStrings() {
		adminSection.DeleteKey(key)
	}
	for k, v := range cfg.Admins {
		adminSection.Key(k).SetValue(v)
	}

	return cfg.file.SaveTo(cfg.Path)
}

// findConfigPath returns the first found default config path if cliPath is empty.
func findConfigPath(cliPath string) string {
	if cliPath != "" {
		return cliPath
	}
	paths := []string{
		"/data/fsm.ini", // Zero-Config Unraid Mapping
		"/fsm.ini", // Legacy root mapped path
		"./fsm.ini",
		filepath.Join(os.Getenv("HOME"), ".config/fsm/fsm.ini"),
		"/etc/fsm/fsm.ini",
	}
	for _, path := range paths {
		if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
			return path
		}
	}

	// Zero-Config Auto-generation
	genPath := "./fsm.ini"
	if stat, err := os.Stat("/data"); err == nil && stat.IsDir() {
		genPath = "/data/fsm.ini" // Generate securely inside the Docker volume
	}

	defaultConfig := `[factorio]
auto_start      = false
config          = /data/config
mods            = /data/mods
bind            = 0.0.0.0:34197
saves           = /data/saves
logs            = /data/logs
downloads       = /data/downloads
server_versions = /data/servers
username        =
token           =

[rcon]
bind     = 0.0.0.0:27015 ; Unraid Optimization: Bind to all interfaces for Docker
password = ChangeMe

[server]
listen = :8888 ; Unraid Optimization: Default to 8888
`
	if err := os.WriteFile(genPath, []byte(defaultConfig), 0644); err != nil {
		log.Println("Could not write Zero-Config default file to", genPath, ":", err)
	} else {
		log.Println("Zero-Config automatically generated default settings at", genPath)
	}

	return genPath
}
