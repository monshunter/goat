package increment

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/monshunter/goat/pkg/config"
)

func TestNewValues(t *testing.T) {
	cfg := &config.Config{
		GoatPackageName: "testpkg",
		AppVersion:      "1.0.0",
		AppName:         "TestApp",
		Race:            true,
	}
	v := NewValues(cfg)

	if v.PackageName != "testpkg" {
		t.Errorf("Expected PackageName to be testpkg, got %s", v.PackageName)
	}
	if v.Version != "1.0.0" {
		t.Errorf("Expected Version to be 1.0.0, got %s", v.Version)
	}
	if v.Name != "TestApp" {
		t.Errorf("Expected Name to be TestApp, got %s", v.Name)
	}
	if !v.Race {
		t.Errorf("Expected Race to be true, got %v", v.Race)
	}
	if len(v.Components) != 0 {
		t.Errorf("Expected empty components list, got %d components", len(v.Components))
	}
	if len(v.TrackIds) != 0 {
		t.Errorf("Expected empty TrackIds list, got %d IDs", len(v.TrackIds))
	}
}

func TestAddComponent(t *testing.T) {
	cfg := &config.Config{
		GoatPackageName: "testpkg",
		AppVersion:      "1.0.0",
		AppName:         "TestApp",
		Race:            true,
	}
	v := NewValues(cfg)

	trackIds := []int{1, 2, 3}
	v.AddComponent(1, "TestComponent", trackIds)

	if len(v.Components) != 1 {
		t.Fatalf("Expected 1 component, got %d", len(v.Components))
	}

	comp := v.Components[0]
	if comp.ID != 1 {
		t.Errorf("Expected component ID to be 1, got %d", comp.ID)
	}
	if comp.Name != "TestComponent" {
		t.Errorf("Expected component name to be TestComponent, got %s", comp.Name)
	}
	if !reflect.DeepEqual(comp.TrackIds, trackIds) {
		t.Errorf("Expected TrackIds to be %v, got %v", trackIds, comp.TrackIds)
	}
}
func TestAddTrackId(t *testing.T) {
	cfg := &config.Config{
		GoatPackageName: "testpkg",
		AppVersion:      "1.0.0",
		AppName:         "TestApp",
		Race:            true,
	}
	v := NewValues(cfg)

	v.AddTrackId(1)
	v.AddTrackId(2)

	if len(v.TrackIds) != 2 {
		t.Fatalf("Expected 2 TrackIds, got %d", len(v.TrackIds))
	}

	if v.TrackIds[0] != 1 || v.TrackIds[1] != 2 {
		t.Errorf("Expected TrackIds to be [1, 2], got %v", v.TrackIds)
	}
}

func TestValidate(t *testing.T) {
	testCases := []struct {
		name      string
		values    *Values
		expectErr bool
	}{
		{
			name: "valid values",
			values: &Values{
				PackageName: "testpkg",
				Name:        "TestApp",
				TrackIds:    []int{1},
			},
			expectErr: false,
		},
		{
			name: "missing package name",
			values: &Values{
				Name:     "TestApp",
				TrackIds: []int{1},
			},
			expectErr: true,
		},
		{
			name: "missing application name",
			values: &Values{
				PackageName: "testpkg",
				TrackIds:    []int{1},
			},
			expectErr: true,
		},
		{
			name: "missing TrackIds",
			values: &Values{
				PackageName: "testpkg",
				Name:        "TestApp",
			},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.values.Validate()
			if tc.expectErr && err == nil {
				t.Errorf("expected validation to fail but it passed")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("expected validation to pass but got error: %v", err)
			}
		})
	}
}

func TestMergeValues(t *testing.T) {
	cfg := &config.Config{
		GoatPackageName: "testpkg",
		AppVersion:      "1.0.0",
		AppName:         "TestApp",
		Race:            true,
	}
	v1 := NewValues(cfg)
	v1.AddComponent(1, "Component1", []int{1, 2})
	v1.AddTrackId(1)
	v1.AddTrackId(2)

	v2 := NewValues(cfg)
	v2.AddComponent(1, "Component1Updated", []int{2, 3})
	v2.AddComponent(2, "Component2", []int{4, 5})
	v2.AddTrackId(2)
	v2.AddTrackId(3)

	v1.MergeValues(v2)

	// Verify TrackIds
	expectedTrackIds := []int{1, 2, 3}
	if !reflect.DeepEqual(v1.TrackIds, expectedTrackIds) {
		t.Errorf("Expected TrackIds after merge to be %v, got %v", expectedTrackIds, v1.TrackIds)
	}

	// Verify Components
	if len(v1.Components) != 2 {
		t.Fatalf("Expected 2 components after merge, got %d", len(v1.Components))
	}

	// Check if first component was updated
	comp1 := v1.Components[0]
	if comp1.Name != "Component1Updated" {
		t.Errorf("Expected component1 name to be updated to Component1Updated, got %s", comp1.Name)
	}
	expectedComp1TrackIds := []int{1, 2, 3}
	if !reflect.DeepEqual(comp1.TrackIds, expectedComp1TrackIds) {
		t.Errorf("Expected component1 TrackIds to be %v, got %v", expectedComp1TrackIds, comp1.TrackIds)
	}

	// Check if second component was added
	comp2 := v1.Components[1]
	if comp2.ID != 2 || comp2.Name != "Component2" {
		t.Errorf("Expected to add component2 (ID=2, Name=Component2), got (ID=%d, Name=%s)", comp2.ID, comp2.Name)
	}
}

func TestClone(t *testing.T) {
	cfg := &config.Config{
		GoatPackageName: "testpkg",
		AppVersion:      "1.0.0",
		AppName:         "TestApp",
		Race:            true,
	}
	v := NewValues(cfg)
	v.AddComponent(1, "Component1", []int{1, 2})
	v.AddTrackId(1)
	v.AddTrackId(2)

	clone := v.Clone()

	// Verify cloned values match original
	if !reflect.DeepEqual(v.TrackIds, clone.TrackIds) {
		t.Errorf("Clone TrackIds failed, original: %v, clone: %v", v.TrackIds, clone.TrackIds)
	}

	if !reflect.DeepEqual(v.Components, clone.Components) {
		t.Errorf("Clone Components failed")
	}

	// Modify original to verify deep copy
	v.TrackIds[0] = 99
	v.Components[0].TrackIds[0] = 99

	if clone.TrackIds[0] == 99 {
		t.Errorf("Clone TrackIds should not be affected by original modification")
	}

	if clone.Components[0].TrackIds[0] == 99 {
		t.Errorf("Clone Component TrackIds should not be affected by original modification")
	}
}

func TestRenderWithCustomTemplate(t *testing.T) {
	cfg := &config.Config{
		GoatPackageName: "testpkg",
		AppVersion:      "1.0.0",
		AppName:         "TestApp",
		Race:            true,
	}
	v := NewValues(cfg)
	v.AddTrackId(1)

	template := "PackageName: {{.PackageName}}, Name: {{.Name}}"

	result, err := v.RenderWithCustomTemplate(template)
	if err != nil {
		t.Fatalf("Rendering failed: %v", err)
	}

	expected := "PackageName: testpkg, Name: TestApp"
	if string(result) != expected {
		t.Errorf("Expected rendering result %s, got %s", expected, string(result))
	}
}
func TestBuildCustomTemplate(t *testing.T) {
	header := "// Header"
	body := "// Body"
	footer := "// Footer"

	template := BuildCustomTemplate(header, body, footer)

	expected := "// Header\n// Body\n// Footer"
	if template != expected {
		t.Errorf("Expected template to be %s, got %s", expected, template)
	}
}
func TestSave(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "render-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		GoatPackageName: "testpkg",
		AppVersion:      "1.0.0",
		AppName:         "TestApp",
		Race:            true,
	}
	v := NewValues(cfg)
	v.AddTrackId(1)

	filePath := filepath.Join(tempDir, "test_output.go")

	// Test save functionality
	err = v.Save(filePath)
	if err != nil {
		t.Fatalf("Failed to save file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("File does not exist after save: %s", filePath)
	}
}
func TestRenderToString(t *testing.T) {
	cfg := &config.Config{
		GoatPackageName: "testpkg",
		AppVersion:      "1.0.0",
		AppName:         "TestApp",
		Race:            true,
	}
	v := NewValues(cfg)
	v.AddTrackId(1)

	// Cannot directly modify the constant Template, use RenderWithCustomTemplate method for testing
	customTemplate := "Package: {{.PackageName}}"
	customResult, err := v.RenderWithCustomTemplate(customTemplate)
	if err != nil {
		t.Fatalf("Rendering with custom template failed: %v", err)
	}

	expectedCustom := "Package: testpkg"
	if string(customResult) != expectedCustom {
		t.Errorf("Expected custom template rendering result %s, got %s", expectedCustom, string(customResult))
	}

	// Test RenderToString method
	result, err := v.RenderToString()
	if err != nil {
		t.Fatalf("Rendering to string failed: %v", err)
	}

	// Only check if the result is a non-empty string since we can't predict the full template output
	if result == "" {
		t.Errorf("RenderToString returned an empty string")
	}
}

func TestRegexpDemo1(t *testing.T) {
	const STR = `
package main

import (
		"encoding/json"
		"fmt"
		"os"

		"github.com/monshunter/ast-practice/pkg/getcomments"
		goat "github.com/monshunter/ast-practice/goat"
)

func main() {
		// +goat:start, do not edit
		// +goat:main, do not edit
		// +goat: generated by increament tracking package, do not edit.
		goat.ServeHTTP(goat.COMPONENT_1)
		// +goat:end
		// 检查参数
		if len(os.Args) != 2 {
			fmt.Println("用法: getcomments <文件路径或代码内容>")
			os.Exit(1)
		}

		input := os.Args[1]

		// 提取注释
		// commentsMap, err := ExtractComments(input)
		// +goat:start, do not edit
		// +goat:generate, do not edit
		goat.Track(goat.TRACK_ID_1)
		// +goat:end
		commentsMap, err := getcomments.ExtractComments(input)
		if err != nil {
			fmt.Printf("提取注释失败: %v\n", err)
			os.Exit(1)
		}
		// +goat:start, do not edit
		// +goat:generate, do not edit
		goat.Track(goat.TRACK_ID_2)
		// +goat:end
		// 输出JSON格式结果
		output, err := json.MarshalIndent(commentsMap, "", "    ")
		if err != nil {
			fmt.Printf("序列化结果失败: %v\n", err)
			os.Exit(1)
		}
		// +goat:start, do not edit
		// +goat:generate, do not edit
		goat.Track(goat.TRACK_ID_3)
		// +goat:end
		fmt.Println(string(output))
}
`
	// 将STR中的goat.Track(goat.TRACK_ID_*) 全部替换成goat.Track(goat.TRACK_ID)
	re := regexp.MustCompile(`goat\.Track\(goat\.TRACK_ID_\d+\)`)
	replaced := re.ReplaceAllString(STR, "goat.Track(goat.TRACK_ID)")

	// 验证替换结果
	expected := `
package main

import (
		"encoding/json"
		"fmt"
		"os"

		"github.com/monshunter/ast-practice/pkg/getcomments"
		goat "github.com/monshunter/ast-practice/goat"
)

func main() {
		// +goat:start, do not edit
		// +goat:main, do not edit
		// +goat: generated by increament tracking package, do not edit.
		goat.ServeHTTP(goat.COMPONENT_1)
		// +goat:end
		// 检查参数
		if len(os.Args) != 2 {
			fmt.Println("用法: getcomments <文件路径或代码内容>")
			os.Exit(1)
		}

		input := os.Args[1]

		// 提取注释
		// commentsMap, err := ExtractComments(input)
		// +goat:start, do not edit
		// +goat:generate, do not edit
		goat.Track(goat.TRACK_ID)
		// +goat:end
		commentsMap, err := getcomments.ExtractComments(input)
		if err != nil {
			fmt.Printf("提取注释失败: %v\n", err)
			os.Exit(1)
		}
		// +goat:start, do not edit
		// +goat:generate, do not edit
		goat.Track(goat.TRACK_ID)
		// +goat:end
		// 输出JSON格式结果
		output, err := json.MarshalIndent(commentsMap, "", "    ")
		if err != nil {
			fmt.Printf("序列化结果失败: %v\n", err)
			os.Exit(1)
		}
		// +goat:start, do not edit
		// +goat:generate, do not edit
		goat.Track(goat.TRACK_ID)
		// +goat:end
		fmt.Println(string(output))
}
`

	if replaced != expected {
		t.Errorf("正则表达式替换结果不符合预期")
	}

	// 验证替换次数
	originalMatches := re.FindAllString(STR, -1)
	if len(originalMatches) != 3 {
		t.Errorf("原始字符串中应该有3个匹配项，但找到了 %d 个", len(originalMatches))
	}

	// 验证替换后不再有带数字的TRACK_ID
	reAfter := regexp.MustCompile(`goat\.Track\(goat\.TRACK_ID_\d+\)`)
	afterMatches := reAfter.FindAllString(replaced, -1)
	if len(afterMatches) != 0 {
		t.Errorf("替换后仍然有 %d 个带数字的TRACK_ID", len(afterMatches))
	}
}
