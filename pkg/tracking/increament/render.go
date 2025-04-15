package increament

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

type Values struct {
	PackageName string
	Version     string
	Name        string
	Components  []Component
	TrackIds    []int
	Race        bool
}

type Component struct {
	ID       int
	Name     string
	TrackIds []int
}

// NewValues creates a new Values instance
func NewValues(packageName, version, name string, race bool) *Values {
	return &Values{
		PackageName: packageName,
		Version:     version,
		Name:        name,
		Components:  make([]Component, 0),
		TrackIds:    make([]int, 0),
		Race:        race,
	}
}

// AddComponent adds a component to the Values
func (v *Values) AddComponent(id int, name string, trackIds []int) {
	v.Components = append(v.Components, Component{
		ID:       id,
		Name:     name,
		TrackIds: trackIds,
	})
}

// AddTrackId adds a track ID to the Values
func (v *Values) AddTrackId(id int) {
	v.TrackIds = append(v.TrackIds, id)
}

// AddTrackIds adds multiple track IDs to the Values
func (v *Values) AddTrackIds(ids []int) {
	v.TrackIds = append(v.TrackIds, ids...)
}

// Validate validates the parameters of the Values
func (v *Values) Validate() error {
	if v.PackageName == "" {
		return fmt.Errorf("package name is required")
	}
	if v.Name == "" {
		return fmt.Errorf("name is required")
	}
	if len(v.TrackIds) == 0 {
		return fmt.Errorf("at least one track ID is required")
	}
	return nil
}

// MergeValues merges the content of another Values object into the current Values
func (v *Values) MergeValues(other *Values) {
	// Merge Components
	for _, comp := range other.Components {
		found := false
		for i, existing := range v.Components {
			if existing.ID == comp.ID {
				// Update existing component
				v.Components[i].Name = comp.Name
				// Merge TrackIds (deduplicate)
				for _, trackID := range comp.TrackIds {
					exists := false
					for _, existingID := range v.Components[i].TrackIds {
						if existingID == trackID {
							exists = true
							break
						}
					}
					if !exists {
						v.Components[i].TrackIds = append(v.Components[i].TrackIds, trackID)
					}
				}
				found = true
				break
			}
		}
		if !found {
			// 添加新组件
			v.Components = append(v.Components, comp)
		}
	}

	// Merge TrackIds (deduplicate)
	for _, id := range other.TrackIds {
		exists := false
		for _, existingID := range v.TrackIds {
			if existingID == id {
				exists = true
				break
			}
		}
		if !exists {
			v.TrackIds = append(v.TrackIds, id)
		}
	}
}

// Clone creates a deep copy of the Values
func (v *Values) Clone() *Values {
	newValues := &Values{
		PackageName: v.PackageName,
		Version:     v.Version,
		Name:        v.Name,
		Race:        v.Race,
		TrackIds:    make([]int, len(v.TrackIds)),
		Components:  make([]Component, len(v.Components)),
	}

	// 复制TrackIds
	copy(newValues.TrackIds, v.TrackIds)

	// Deep copy Components
	for i, comp := range v.Components {
		newValues.Components[i] = Component{
			ID:       comp.ID,
			Name:     comp.Name,
			TrackIds: make([]int, len(comp.TrackIds)),
		}
		copy(newValues.Components[i].TrackIds, comp.TrackIds)
	}

	return newValues
}

func (v *Values) Render() ([]byte, error) {
	tmpl, err := template.New("").Parse(Template)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderWithCustomTemplate renders data with a custom template
func (v *Values) RenderWithCustomTemplate(customTemplate string) ([]byte, error) {
	tmpl, err := template.New("custom").Parse(customTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse custom template: %w", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, v); err != nil {
		return nil, fmt.Errorf("failed to execute custom template: %w", err)
	}
	return buf.Bytes(), nil
}

// Save saves the rendered result to a file
func (v *Values) Save(outputPath string) error {
	if err := v.Validate(); err != nil {
		return err
	}

	data, err := v.Render()
	if err != nil {
		return err
	}

	// Ensure the directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0644)
}

func (v *Values) Remove(outputPath string) error {
	return os.Remove(outputPath)
}

// SaveWithCustomTemplate renders with a custom template and saves to a file
func (v *Values) SaveWithCustomTemplate(outputPath, customTemplate string) error {
	if err := v.Validate(); err != nil {
		return err
	}

	data, err := v.RenderWithCustomTemplate(customTemplate)
	if err != nil {
		return err
	}

	// Ensure the directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0644)
}

// RenderToString renders the result to a string
func (v *Values) RenderToString() (string, error) {
	data, err := v.Render()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// BuildCustomTemplate builds a complete template from template fragments
func BuildCustomTemplate(header, body, footer string) string {
	var buf bytes.Buffer

	if header != "" {
		buf.WriteString(header)
		buf.WriteString("\n")
	}

	if body != "" {
		buf.WriteString(body)
		buf.WriteString("\n")
	}

	if footer != "" {
		buf.WriteString(footer)
	}

	return buf.String()
}
