package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProjectType(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(dir string) error
		expected ProjectType
	}{
		{
			name: "Next.js project",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies":{"next":"^14.0.0"}}`), 0644)
			},
			expected: ProjectNextJS,
		},
		{
			name: "React project",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies":{"react":"^18.0.0"}}`), 0644)
			},
			expected: ProjectReact,
		},
		{
			name: "Vue project",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies":{"vue":"^3.0.0"}}`), 0644)
			},
			expected: ProjectVue,
		},
		{
			name: "Go project",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/test\n\ngo 1.21\n"), 0644)
			},
			expected: ProjectGo,
		},
		{
			name: "Python project",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("flask==2.0.0\n"), 0644)
			},
			expected: ProjectFlask, // Detector is smart enough to detect Flask
		},
		{
			name: "Django project",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("Django==4.2.0\n"), 0644)
			},
			expected: ProjectDjango,
		},
		{
			name: "Flask project",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("Flask==2.3.0\n"), 0644)
			},
			expected: ProjectFlask,
		},
		{
			name: "Rails project",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "Gemfile"), []byte("source 'https://rubygems.org'\ngem 'rails', '~> 7.0'\n"), 0644)
			},
			expected: ProjectRails,
		},
		{
			name: "Java project",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "pom.xml"), []byte("<project></project>"), 0644)
			},
			expected: ProjectJava,
		},
		{
			name: "Rust project",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"test\"\nversion = \"0.1.0\"\n"), 0644)
			},
			expected: ProjectRust,
		},
		{
			name: "Static HTML",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "index.html"), []byte("<!DOCTYPE html><html></html>"), 0644)
			},
			expected: ProjectStatic,
		},
		{
			name: "Docker project",
			setup: func(dir string) error {
				return os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM alpine\n"), 0644)
			},
			expected: ProjectDocker,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir := t.TempDir()

			// Setup test files
			if err := tt.setup(tmpDir); err != nil {
				t.Fatalf("setup failed: %v", err)
			}

			// Change to temp directory
			oldDir, _ := os.Getwd()
			defer os.Chdir(oldDir)
			os.Chdir(tmpDir)

			// Detect project
			project, err := DetectProject()
			if err != nil {
				t.Fatalf("DetectProject failed: %v", err)
			}

			if project.Type != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, project.Type)
			}
		})
	}
}

func TestParseEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := `# Comment line
DATABASE_URL=postgresql://localhost/db
API_KEY=abc123
EMPTY_VAR=
QUOTED_VAR="hello world"
SINGLE_QUOTED='test'
`

	if err := os.WriteFile(envFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write .env: %v", err)
	}

	vars, err := ParseEnvFile(envFile)
	if err != nil {
		t.Fatalf("ParseEnvFile failed: %v", err)
	}

	expected := map[string]string{
		"DATABASE_URL":  "postgresql://localhost/db",
		"API_KEY":       "abc123",
		"EMPTY_VAR":     "",
		"QUOTED_VAR":    "hello world",
		"SINGLE_QUOTED": "test",
	}

	if len(vars) != len(expected) {
		t.Errorf("expected %d vars, got %d", len(expected), len(vars))
	}

	for _, v := range vars {
		if expectedVal, ok := expected[v.Key]; ok {
			if v.Value != expectedVal {
				t.Errorf("var %s: expected %q, got %q", v.Key, expectedVal, v.Value)
			}
		}
	}
}

func TestGenerateDockerfile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		projectType ProjectType
		contains    []string
	}{
		{
			name:        "Go Dockerfile",
			projectType: ProjectGo,
			contains:    []string{"FROM golang:", "CGO_ENABLED=0", "EXPOSE 8080"},
		},
		{
			name:        "Next.js Dockerfile",
			projectType: ProjectNextJS,
			contains:    []string{"FROM node:", "npm run build", "EXPOSE 3000"},
		},
		{
			name:        "Flask Dockerfile",
			projectType: ProjectFlask,
			contains:    []string{"FROM python:", "gunicorn", "EXPOSE 8000"},
		},
		{
			name:        "Rails Dockerfile",
			projectType: ProjectRails,
			contains:    []string{"FROM ruby:", "bundle install", "EXPOSE 3000"},
		},
		{
			name:        "Static Dockerfile",
			projectType: ProjectStatic,
			contains:    []string{"FROM nginx:", "EXPOSE 80"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project := &ProjectInfo{
				Type: tt.projectType,
				Port: getDefaultPort(tt.projectType),
			}

			oldDir, _ := os.Getwd()
			defer os.Chdir(oldDir)
			os.Chdir(tmpDir)

			if err := GenerateDockerfile(project); err != nil {
				t.Fatalf("GenerateDockerfile failed: %v", err)
			}

			content, err := os.ReadFile("Dockerfile")
			if err != nil {
				t.Fatalf("failed to read Dockerfile: %v", err)
			}

			for _, str := range tt.contains {
				if !contains(string(content), str) {
					t.Errorf("Dockerfile should contain %q", str)
				}
			}

			os.Remove("Dockerfile")
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func getDefaultPort(pt ProjectType) string {
	switch pt {
	case ProjectNextJS, ProjectReact, ProjectRails:
		return "3000"
	case ProjectGo, ProjectJava, ProjectRust:
		return "8080"
	case ProjectFlask, ProjectDjango, ProjectPython, ProjectStatic:
		return "8000"
	default:
		return "8080"
	}
}
