package config

import (
	"path/filepath"
	"testing"
)

func TestPrinterConfigModeIsValid(t *testing.T) {
	tests := []struct {
		name string
		mode PrinterConfigMode
		want bool
	}{
		{"empty mode", PrinterConfigModeNone, true},
		{"use spaces", PrinterConfigModeUseSpaces, true},
		{"tab indent", PrinterConfigModeTabIndent, true},
		{"source pos", PrinterConfigModeSourcePos, true},
		{"raw format", PrinterConfigModeRawFormat, true},
		{"invalid mode", "invalidMode", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.IsValid(); got != tt.want {
				t.Errorf("PrinterConfigMode.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrinterConfigModeMode(t *testing.T) {
	tests := []struct {
		name string
		mode PrinterConfigMode
	}{
		{"empty mode", PrinterConfigModeNone},
		{"use spaces", PrinterConfigModeUseSpaces},
		{"tab indent", PrinterConfigModeTabIndent},
		{"source pos", PrinterConfigModeSourcePos},
		{"raw format", PrinterConfigModeRawFormat},
		{"invalid mode", "invalidMode"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just ensure it doesn't panic
			_ = tt.mode.Mode()
		})
	}
}

func TestToGranularity(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Granularity
		wantErr bool
	}{
		{"line", GranularityLineStr, GranularityLine, false},
		{"patch", GranularityPatchStr, GranularityPatch, false},
		{"func", GranularityFuncStr, GranularityFunc, false},
		{"scope", GranularityScopeStr, GranularityScope, false},
		{"invalid", "invalid", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToGranularity(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToGranularity() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToGranularity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGranularityIsValid(t *testing.T) {
	tests := []struct {
		name        string
		granularity Granularity
		want        bool
	}{
		{"line", GranularityLine, true},
		{"patch", GranularityPatch, true},
		{"func", GranularityFunc, true},
		{"scope", GranularityScope, true},
		{"invalid", Granularity(0), false},
		{"invalid high", Granularity(100), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.granularity.IsValid(); got != tt.want {
				t.Errorf("Granularity.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGranularityString(t *testing.T) {
	tests := []struct {
		name        string
		granularity Granularity
	}{
		{"line", GranularityLine},
		{"patch", GranularityPatch},
		{"func", GranularityFunc},
		{"scope", GranularityScope},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just ensure it doesn't panic
			_ = tt.granularity.String()
		})
	}
}

func TestGranularityInt(t *testing.T) {
	tests := []struct {
		name        string
		granularity Granularity
	}{
		{"line", GranularityLine},
		{"patch", GranularityPatch},
		{"func", GranularityFunc},
		{"scope", GranularityScope},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just ensure it doesn't panic
			_ = tt.granularity.Int()
		})
	}
}

func TestGranularityChecks(t *testing.T) {
	tests := []struct {
		name        string
		granularity Granularity
		isLine      bool
		isPatch     bool
		isFunc      bool
		isScope     bool
	}{
		{"line", GranularityLine, true, false, false, false},
		{"patch", GranularityPatch, false, true, false, false},
		{"func", GranularityFunc, false, false, true, false},
		{"scope", GranularityScope, false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.granularity.IsLine(); got != tt.isLine {
				t.Errorf("Granularity.IsLine() = %v, want %v", got, tt.isLine)
			}
			if got := tt.granularity.IsPatch(); got != tt.isPatch {
				t.Errorf("Granularity.IsPatch() = %v, want %v", got, tt.isPatch)
			}
			if got := tt.granularity.IsFunc(); got != tt.isFunc {
				t.Errorf("Granularity.IsFunc() = %v, want %v", got, tt.isFunc)
			}
			if got := tt.granularity.IsScope(); got != tt.isScope {
				t.Errorf("Granularity.IsScope() = %v, want %v", got, tt.isScope)
			}
		})
	}
}

func TestDataTypeString(t *testing.T) {
	tests := []struct {
		name     string
		dataType DataType
		want     string
	}{
		{"truth", DataTypeTruth, "truth"},
		{"count", DataTypeCount, "count"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dataType.String(); got != tt.want {
				t.Errorf("DataType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDataTypeIsValid(t *testing.T) {
	tests := []struct {
		name     string
		dataType DataType
		want     bool
	}{
		{"truth", DataTypeTruth, true},
		{"count", DataTypeCount, true},
		{"invalid", DataType(0), false},
		{"invalid high", DataType(100), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dataType.IsValid(); got != tt.want {
				t.Errorf("DataType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDataTypeInt(t *testing.T) {
	tests := []struct {
		name     string
		dataType DataType
		want     int
	}{
		{"truth", DataTypeTruth, 1},
		{"count", DataTypeCount, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.dataType.Int(); got != tt.want {
				t.Errorf("DataType.Int() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDataType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    DataType
		wantErr bool
	}{
		{"truth", "truth", DataTypeTruth, false},
		{"count", "count", DataTypeCount, false},
		{"invalid", "invalid", DataTypeTruth, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDataType(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDataType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetDataType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	// Create a minimal valid config with AppVersion set to avoid git operations
	validConfig := &Config{
		DiffPrecision: 1,
		AppVersion:    "test-version", // Set this to avoid git operations
	}

	// Test validation of basic properties
	if err := validConfig.Validate(); err != nil {
		t.Errorf("Config.Validate() error = %v, wantErr false", err)
	}

	// Test with invalid granularity
	invalidGranularity := &Config{
		Granularity:   "invalid",
		DiffPrecision: 1,
		AppVersion:    "test-version",
	}
	if err := invalidGranularity.Validate(); err == nil {
		t.Errorf("Config.Validate() with invalid granularity error = nil, wantErr true")
	}

	// Test with invalid diff precision
	invalidDiffPrecision := &Config{
		DiffPrecision: 0,
		AppVersion:    "test-version",
	}
	if err := invalidDiffPrecision.Validate(); err == nil {
		t.Errorf("Config.Validate() with invalid diff precision error = nil, wantErr true")
	}

	// Test with invalid printer config mode
	invalidPrinterMode := &Config{
		DiffPrecision:     1,
		PrinterConfigMode: []PrinterConfigMode{"invalid"},
		AppVersion:        "test-version",
	}
	if err := invalidPrinterMode.Validate(); err == nil {
		t.Errorf("Config.Validate() with invalid printer config mode error = nil, wantErr true")
	}

	// Test with invalid data type
	invalidDataType := &Config{
		DiffPrecision: 1,
		DataType:      "invalid",
		AppVersion:    "test-version",
	}
	if err := invalidDataType.Validate(); err == nil {
		t.Errorf("Config.Validate() with invalid data type error = nil, wantErr true")
	}
}

func TestConfigGetGranularity(t *testing.T) {
	tests := []struct {
		name        string
		granularity string
		want        Granularity
	}{
		{"line", GranularityLineStr, GranularityLine},
		{"patch", GranularityPatchStr, GranularityPatch},
		{"func", GranularityFuncStr, GranularityFunc},
		{"scope", GranularityScopeStr, GranularityScope},
		{"invalid", "invalid", GranularityPatch}, // Default to patch for invalid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Granularity: tt.granularity,
			}
			if got := c.GetGranularity(); got != tt.want {
				t.Errorf("Config.GetGranularity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigIsMainEntry(t *testing.T) {
	tests := []struct {
		name        string
		mainEntries []string
		entry       string
		want        bool
	}{
		{"wildcard", []string{"*"}, "any", true},
		{"exact match", []string{"main", "cmd"}, "main", true},
		{"no match", []string{"main", "cmd"}, "other", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				MainEntries: tt.mainEntries,
			}
			if got := c.IsMainEntry(tt.entry); got != tt.want {
				t.Errorf("Config.IsMainEntry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigGoatGeneratedFile(t *testing.T) {
	c := &Config{
		GoatPackagePath: "path/to/goat",
	}
	expected := filepath.Join("path/to/goat", goatGeneratedFile)
	if got := c.GoatGeneratedFile(); got != expected {
		t.Errorf("Config.GoatGeneratedFile() = %v, want %v", got, expected)
	}
}

func TestConfigPrinterConfig(t *testing.T) {
	c := &Config{
		PrinterConfigMode:     []PrinterConfigMode{PrinterConfigModeUseSpaces},
		PrinterConfigTabwidth: 4,
		PrinterConfigIndent:   2,
	}

	// First call should create a new printer config
	cfg1 := c.PrinterConfig()
	if cfg1 == nil {
		t.Errorf("Config.PrinterConfig() = nil, want non-nil")
	}

	// Second call should return the cached config
	cfg2 := c.PrinterConfig()
	if cfg1 != cfg2 {
		t.Errorf("Config.PrinterConfig() returned different instances on subsequent calls")
	}
}

func TestConfigGetDataType(t *testing.T) {
	tests := []struct {
		name     string
		dataType string
		want     DataType
	}{
		{"truth", "truth", DataTypeTruth},
		{"count", "count", DataTypeCount},
		{"invalid", "invalid", DataTypeTruth}, // Default to truth for invalid
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				DataType: tt.dataType,
			}
			if got := c.GetDataType(); got != tt.want {
				t.Errorf("Config.GetDataType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigIsNewRepository(t *testing.T) {
	tests := []struct {
		name      string
		oldBranch string
		want      bool
	}{
		{"new repo", "INIT", true},
		{"existing repo", "main", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				OldBranch: tt.oldBranch,
			}
			if got := c.IsNewRepository(); got != tt.want {
				t.Errorf("Config.IsNewRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}
