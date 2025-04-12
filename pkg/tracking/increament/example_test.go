//go:build example
// +build example

package increament_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/monshunter/goat/pkg/tracking/increament"
)

// This is an example of how to use the Values type to generate tracking code
func Example() {
	// Create a new Values instance
	values := increament.NewValues(
		"myapp",      // Package name
		"1.0.0",      // Version
		"ExampleApp", // Application name
		true,         // Enable race condition protection
	)

	// Add track IDs
	values.AddTrackId(100)
	values.AddTrackId(101)
	values.AddTrackId(102)

	// Add components and their track IDs
	values.AddComponent(1, "LoginComponent", []int{100, 101})
	values.AddComponent(2, "DashboardComponent", []int{102})

	// Validate the values
	if err := values.Validate(); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	// Render to string
	result, err := values.RenderToString()
	if err != nil {
		fmt.Printf("Rendering failed: %v\n", err)
		return
	}

	fmt.Printf("Successfully generated code, length is %d bytes\n", len(result))

	// Save to temporary file
	tempDir, err := os.MkdirTemp("", "example")
	if err != nil {
		fmt.Printf("Failed to create temp directory: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	outputPath := filepath.Join(tempDir, "track.go")
	if err := values.Save(outputPath); err != nil {
		fmt.Printf("Failed to save file: %v\n", err)
		return
	}

	fmt.Printf("Successfully saved to file: %s\n", outputPath)

	// Use custom template
	customTemplate := `
// Custom template example
package {{.PackageName}}

// Track ID constants
const (
	{{range .TrackIds}}
	TRACK_ID_{{.}} = {{.}}
	{{end}}
)

// Component information
{{range .Components}}
// {{.Name}} component (ID={{.ID}})
var {{.Name}}_TrackIds = []int{
	{{range .TrackIds}}
	TRACK_ID_{{.}},
	{{end}}
}
{{end}}
`

	customResult, err := values.RenderWithCustomTemplate(customTemplate)
	if err != nil {
		fmt.Printf("Custom template rendering failed: %v\n", err)
		return
	}

	fmt.Printf("Custom template rendering succeeded, content length is %d bytes\n", len(customResult))

	// Output: Successfully generated code, code length is xxx bytes
	//       Successfully saved to file: /tmp/xxx/track.go
	//       Custom template rendering succeeded, content length is xxx bytes
}
