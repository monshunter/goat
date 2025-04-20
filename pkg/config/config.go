package config

import (
	"fmt"
	"go/printer"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"golang.org/x/mod/modfile"
	"gopkg.in/yaml.v3"
)

const goatGeneratedFile = "goat_generated.go"

const (
	// Track generate comment, which is used to mark the generate of the track
	TrackGenerateComment = "// +goat:generate"
	// Track tips comment, which is used to mark the tips of the track
	TrackTipsComment = "// +goat:tips: do not edit the block between the +goat comments"
	// Track delete comment, which is used to mark the delete of the track
	// goat fix will try to delete codes
	TrackDeleteComment = "// +goat:delete"
	// Track import comment, which is used to mark the import of the track
	TrackImportComment = "// +goat:import"
	// Track insert comment, which is used to mark to add by human
	// goat fix will try to insert codes the goat fix will try to insert codes the goat fix will try to insert codes the
	TrackInsertComment = "// +goat:insert"
	// Track main entry comment, which is used to mark the main entry of the track
	TrackMainEntryComment = "// +goat:main"
	// Track end comment, which is used to mark the end of the track
	TrackEndComment = "// +goat:end"
	// Track user comment, which is used to mark the track is user defined
	TrackUserComment = "// +goat:user"
)

var (
	// Track insert regexp, which is used to match the insert comment
	TrackInsertRegexp = regexp.MustCompile(regexp.QuoteMeta(TrackInsertComment))
	// Track generate end regexp, which is used to match the generate end comment
	TrackGenerateEndRegexp = regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(TrackGenerateComment) +
		`[^\n]` + `*\n(?:.*\n)*?\s*` + regexp.QuoteMeta(TrackEndComment) + `[^\n]*\n`)
	// Track delete end regexp, which is used to match the delete end comment
	TrackDeleteEndRegexp = regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(TrackDeleteComment) +
		`[^\n]` + `*\n(?:.*\n)*?\s*` + regexp.QuoteMeta(TrackEndComment) + `[^\n]*\n`)
	// Track main entry end regexp, which is used to match the main entry end comment
	TrackMainEntryEndRegexp = regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(TrackMainEntryComment) +
		`[^\n]` + `*\n(?:.*\n)*?\s*` + regexp.QuoteMeta(TrackEndComment) + `[^\n]*\n`)
	// Track user end regexp, which is used to match the user end comment
	TrackUserEndRegexp = regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(TrackUserComment) +
		`[^\n]` + `*\n(?:.*\n)*?\s*` + regexp.QuoteMeta(TrackEndComment) + `[^\n]*\n`)
)

type Granularity int

const (
	_ Granularity = iota
	GranularityLine
	GranularityBlock
	GranularityScope
	GranularityFunc
)

const (
	GranularityLineStr  = "line"
	GranularityBlockStr = "block"
	GranularityScopeStr = "scope"
	GranularityFuncStr  = "func"
)

// PrinterConfigMode is the mode of the printer config
type PrinterConfigMode string

// PrinterConfigModeNone is the none mode of the printer config
// default  useSpaces | tabIndent
const (
	PrinterConfigModeNone PrinterConfigMode = ""
	// same as printer.UseSpaces
	PrinterConfigModeUseSpaces PrinterConfigMode = "useSpaces"
	// same as printer.TabIndent
	PrinterConfigModeTabIndent PrinterConfigMode = "tabIndent"
	// same as printer.SourcePos
	PrinterConfigModeSourcePos PrinterConfigMode = "sourcePos"
	// same as printer.RawFormat
	PrinterConfigModeRawFormat PrinterConfigMode = "rawFormat"
)

// IsValid checks if the printer config mode is valid
func (p PrinterConfigMode) IsValid() bool {
	switch p {
	case PrinterConfigModeNone, PrinterConfigModeUseSpaces, PrinterConfigModeTabIndent,
		PrinterConfigModeSourcePos, PrinterConfigModeRawFormat:
		return true
	default:
		return false
	}
}

// Mode returns the printer mode
func (p PrinterConfigMode) Mode() printer.Mode {
	switch p {
	case PrinterConfigModeUseSpaces:
		return printer.UseSpaces
	case PrinterConfigModeTabIndent:
		return printer.TabIndent
	case PrinterConfigModeSourcePos:
		return printer.SourcePos
	case PrinterConfigModeRawFormat:
		return printer.RawFormat
	default:
		return printer.Mode(0)
	}
}

// ToGranularity converts a string to a granularity
func ToGranularity(s string) (Granularity, error) {
	switch s {
	case GranularityLineStr:
		return GranularityLine, nil
	case GranularityBlockStr:
		return GranularityBlock, nil
	case GranularityFuncStr:
		return GranularityFunc, nil
	case GranularityScopeStr:
		return GranularityScope, nil
	default:
		return 0, fmt.Errorf("invalid granularity: %s", s)
	}
}

// IsValid checks if the granularity is valid
func (g Granularity) IsValid() bool {
	switch g {
	case GranularityLine, GranularityBlock, GranularityFunc, GranularityScope:
		return true
	default:
		return false
	}
}

// String returns the string representation of the granularity
func (g Granularity) String() string {
	return []string{GranularityLineStr, GranularityBlockStr, GranularityFuncStr, GranularityScopeStr}[g-1]
}

// Int returns the integer representation of the granularity
func (g Granularity) Int() int {
	return int(g)
}

// IsLine checks if the granularity is line
func (g Granularity) IsLine() bool {
	return g == GranularityLine
}

// IsBlock checks if the granularity is block
func (g Granularity) IsBlock() bool {
	return g == GranularityBlock
}

// IsFunc checks if the granularity is func
func (g Granularity) IsFunc() bool {
	return g == GranularityFunc
}

// IsScope checks if the granularity is scope
func (g Granularity) IsScope() bool {
	return g == GranularityScope
}

// DataType is the type of the data
type DataType int

// DataTypeTruth is the truth type
const (
	_ DataType = iota
	// DataTypeTruth is the truth type
	DataTypeTruth // default
	// DataTypeCount is the count type
	DataTypeCount
	// DataTypeAverage is the average type
	DataTypeAverage
)

// String returns the string representation of the data type
func (d DataType) String() string {
	return dataTypeNames[d]
}

// IsValid checks if the data type is valid
func (d DataType) IsValid() bool {
	switch d {
	case DataTypeTruth, DataTypeCount, DataTypeAverage:
		return true
	default:
		return false
	}
}

// dataTypeNames is the names of the data types
var dataTypeNames = []string{
	DataTypeTruth:   "truth",
	DataTypeCount:   "count",
	DataTypeAverage: "average",
}

// Int returns the integer representation of the data type
func (d DataType) Int() int {
	return int(d)
}

func GetDataType(s string) (DataType, error) {
	switch s {
	case "truth":
		return DataTypeTruth, nil
	case "count":
		return DataTypeCount, nil
	case "average":
		return DataTypeAverage, nil
	default:
		return DataTypeTruth, fmt.Errorf("invalid data type: %s", s)
	}
}

// Config configuration struct

// Config configuration struct
type Config struct {
	// App name
	AppName string `yaml:"appName"` // goat
	// App version
	AppVersion string `yaml:"appVersion"` // 1.0.0
	// Stable branch name
	StableBranch string `yaml:"stableBranch"` // commit hash or branch name or tag name or .
	// Publish branch name
	PublishBranch string `yaml:"publishBranch"` // commit hash or branch name or tag name or .
	// Files or directories to ignore
	Ignores []string `yaml:"ignores"`
	// Goat package name
	GoatPackageName string `yaml:"goatPackageName"`
	// Goat package alias
	GoatPackageAlias string `yaml:"goatPackageAlias"`
	// Goat package path
	GoatPackagePath string `yaml:"goatPackagePath"`
	// Granularity
	Granularity string `yaml:"granularity"` // line, block, scope, func
	// Diff precision
	DiffPrecision int `yaml:"diffPrecision"` // 1~2, 3&4 is not supported
	// Threads
	Threads int `yaml:"threads"` // 1~128
	// Race
	Race bool `yaml:"race"` // true, false
	// Clone branch
	CloneBranch bool `yaml:"cloneBranch"` // true, false
	// Main packages to track
	MainEntries []string `yaml:"mainEntries"`
	// Printer config
	// PrinterConfigMode is the mode of the printer config
	// default  useSpaces | tabIndent
	PrinterConfigMode []PrinterConfigMode `yaml:"printerConfigMode"` // default: useSpaces | tabIndent
	// PrinterConfigTabwidth is the tab width of the printer config
	PrinterConfigTabwidth int `yaml:"printerConfigTabwidth"` // default: 8
	// PrinterConfigIndent is the indent of the printer config
	PrinterConfigIndent int `yaml:"printerConfigIndent"` // default: 0
	// printerConfig is the printer config
	printerConfig *printer.Config `yaml:"-"`
	// Data type
	DataType string `yaml:"dataType"` // default: truth
}

// Validate validates the config
func (c *Config) Validate() error {
	if c.Granularity == "" {
		c.Granularity = GranularityBlockStr
	}
	_, err := ToGranularity(c.Granularity)
	if err != nil {
		return fmt.Errorf("invalid granularity: %w", err)
	}

	if c.DiffPrecision < 1 || c.DiffPrecision > 2 {
		return fmt.Errorf("invalid diff precision: %d", c.DiffPrecision)
	}

	if c.Threads <= 0 {
		c.Threads = runtime.NumCPU()
	}

	if c.StableBranch == "" {
		c.StableBranch = "main"
	}

	if c.PublishBranch == "" {
		c.PublishBranch = "HEAD"
	}

	if c.Ignores == nil {
		c.Ignores = []string{".git", ".gitignore", ".DS_Store", ".idea", ".vscode", ".venv", "vendor", "testdata", "node_modules"}
	}

	if c.MainEntries == nil {
		c.MainEntries = []string{"*"}
	}

	if c.GoatPackageName == "" {
		c.GoatPackageName = "goat"
	}

	if c.GoatPackageAlias == "" {
		c.GoatPackageAlias = "goat"
	}

	if c.GoatPackagePath == "" {
		c.GoatPackagePath = "goat"
	}

	if c.PrinterConfigMode == nil {
		c.PrinterConfigMode = []PrinterConfigMode{PrinterConfigModeUseSpaces, PrinterConfigModeTabIndent}
	}

	for _, mode := range c.PrinterConfigMode {
		if !mode.IsValid() {
			return fmt.Errorf("invalid printer config mode: %s", mode)
		}
	}

	if c.PrinterConfigIndent < 0 {
		c.PrinterConfigIndent = 0
	}

	if c.PrinterConfigTabwidth < 1 {
		c.PrinterConfigTabwidth = 8
	}

	if c.DataType == "" {
		c.DataType = "truth"
	}

	dt, err := GetDataType(c.DataType)
	if err != nil {
		return fmt.Errorf("invalid data type: %w", err)
	}
	c.DataType = dt.String()

	// ignore goat_generated.go
	goatFile := c.GoatGeneratedFile()
	found := false
	for _, file := range c.Ignores {
		if file == goatFile {
			found = true
			break
		}
	}
	if !found {
		c.Ignores = append(c.Ignores, goatFile)
	}

	return nil
}

// GetGranularity returns the granularity
func (c *Config) GetGranularity() Granularity {
	granularity, err := ToGranularity(c.Granularity)
	if err != nil {
		return GranularityBlock
	}
	return granularity
}

// IsMainEntry checks if the entry is a main entry
func (c *Config) IsMainEntry(entry string) bool {
	for _, mainEntry := range c.MainEntries {
		if mainEntry == "*" || mainEntry == entry {
			return true
		}
	}
	return false
}

// GoatGeneratedFile returns the goat generated file path
func (c *Config) GoatGeneratedFile() string {
	return filepath.Join(c.GoatPackagePath, goatGeneratedFile)
}

func (c *Config) PrinterConfig() *printer.Config {
	if c.printerConfig != nil {
		return c.printerConfig
	}
	mode := printer.Mode(0)
	for _, m := range c.PrinterConfigMode {
		mode |= m.Mode()
	}
	cfg := &printer.Config{
		Mode:     mode,
		Tabwidth: c.PrinterConfigTabwidth,
		Indent:   c.PrinterConfigIndent,
	}
	c.printerConfig = cfg
	return cfg
}

func (c *Config) GetDataType() DataType {
	dt, err := GetDataType(c.DataType)
	if err != nil {
		return DataTypeTruth
	}
	return dt
}

// LoadConfig loads configuration from file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}
	return &config, nil
}

// InitWithConfig initializes configuration with a Config struct
func InitWithConfig(filename string, cfg *Config) error {
	// parse config template
	tmpl, err := template.New("config").Parse(CONFIG_TEMPLATE)
	if err != nil {
		return fmt.Errorf("failed to parse config template: %w", err)
	}

	// create config file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// execute template
	if err := tmpl.Execute(file, cfg); err != nil {
		return fmt.Errorf("failed to execute config template: %w", err)
	}

	return nil
}

// GetGoModuleName gets the module name from the go.mod file
func GoModuleName() string {
	modFilePath := "go.mod"
	content, err := os.ReadFile(modFilePath)
	if err != nil {
		panic(fmt.Sprintf("failed to read go.mod file: %v", err))
	}
	modFile, err := modfile.Parse(modFilePath, content, nil)
	if err != nil {
		panic(fmt.Sprintf("failed to parse go.mod file: %v", err))
	}
	return modFile.Module.Mod.Path
}
