package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ReviewFinding represents a single issue found during code review
type ReviewFinding struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	Severity    string `json:"severity"` // HIGH, MEDIUM, LOW
	Category    string `json:"category"` // Bug, Security, Performance, Logic, Style
	Title       string `json:"title"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion"`
	CodeSnippet string `json:"code_snippet"`
}

// ReviewResult holds the full review output
type ReviewResult struct {
	ProjectName  string          `json:"project_name"`
	ProjectType  string          `json:"project_type"`
	FilesScanned int             `json:"files_scanned"`
	TotalLines   int             `json:"total_lines"`
	Findings     []ReviewFinding `json:"findings"`
	ReviewedAt   string          `json:"reviewed_at"`
	Summary      string          `json:"summary"`
	Duration     string          `json:"duration"`
}

// ReviewConfig holds review configuration
type ReviewConfig struct {
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url"`
	Model    string `json:"model"`
	FreeMode bool   `json:"free_mode"`
}

// DefaultReviewConfig returns the default configuration for free period
func DefaultReviewConfig() *ReviewConfig {
	// Check for user-provided API key first
	apiKey := os.Getenv("SHIPMATE_AI_KEY")
	if apiKey == "" {
		// Use free tier OpenRouter models during free period
		apiKey = "sk-or-v1-free-tier" // Placeholder — real key from env
	}

	baseURL := os.Getenv("SHIPMATE_AI_URL")
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1/chat/completions"
	}

	model := os.Getenv("SHIPMATE_AI_MODEL")
	if model == "" {
		model = "deepseek/deepseek-chat-v3-0324:free"
	}

	return &ReviewConfig{
		APIKey:   apiKey,
		BaseURL:  baseURL,
		Model:    model,
		FreeMode: true, // Free during beta period
	}
}

// ReviewCode performs a code review on the current project
func ReviewCode(project *ProjectInfo) (*ReviewResult, error) {
	start := time.Now()
	config := DefaultReviewConfig()

	// Check auth for paid mode (bypassed during free period)
	if !config.FreeMode {
		token, err := GetToken("shipmate")
		if err != nil || token == "" {
			return nil, fmt.Errorf("code review requires a Shipmate Cloud account.\n  Run: shipmate login shipmate\n  Or visit: https://myshipmate.cc/pricing")
		}
	}

	fmt.Println("   📖 Scanning project files...")

	// Collect code files
	files, err := collectReviewFiles(project)
	if err != nil {
		return nil, fmt.Errorf("file collection failed: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no code files found to review")
	}

	totalLines := 0
	for _, f := range files {
		totalLines += f.lines
	}

	fmt.Printf("   ✓ Found %d files (%d lines of code)\n", len(files), totalLines)
	fmt.Println("   🤖 Sending to AI reviewer...")

	// Build the review prompt with all code
	prompt := buildReviewPrompt(project, files)

	// Send to AI for review
	findings, summary, err := performAIReview(config, prompt)
	if err != nil {
		return nil, fmt.Errorf("AI review failed: %w", err)
	}

	elapsed := time.Since(start)

	result := &ReviewResult{
		ProjectName:  project.Name,
		ProjectType:  string(project.Type),
		FilesScanned: len(files),
		TotalLines:   totalLines,
		Findings:     findings,
		ReviewedAt:   time.Now().Format(time.RFC3339),
		Summary:      summary,
		Duration:     elapsed.Round(time.Millisecond).String(),
	}

	return result, nil
}

// ReviewFile represents a file being reviewed
type ReviewFile struct {
	path    string
	relPath string
	content string
	lines   int
}

// collectReviewFiles gathers code files for review
func collectReviewFiles(project *ProjectInfo) ([]ReviewFile, error) {
	var files []ReviewFile

	// Extensions to review (skip binaries, configs, etc.)
	reviewExts := map[string]bool{
		".go": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
		".py": true, ".rb": true, ".java": true, ".kt": true, ".rs": true,
		".cs": true, ".php": true, ".ex": true, ".exs": true, ".dart": true,
		".vue": true, ".svelte": true, ".css": true, ".scss": true,
		".sql": true, ".sh": true, ".bash": true,
	}

	// Directories to skip
	skipDirs := map[string]bool{
		"node_modules": true, ".git": true, ".next": true, ".nuxt": true,
		"dist": true, "build": true, "vendor": true, "__pycache__": true,
		".venv": true, "venv": true, "target": true, ".gradle": true,
		".idea": true, ".vscode": true, "coverage": true, ".cache": true,
		"out": true, ".output": true,
	}

	err := filepath.Walk(project.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			if skipDirs[info.Name()] || strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip large files (>50KB)
		if info.Size() > 50*1024 {
			return nil
		}

		ext := filepath.Ext(info.Name())
		if !reviewExts[ext] {
			return nil
		}

		relPath, _ := filepath.Rel(project.Path, path)

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		lines := strings.Count(string(content), "\n") + 1

		files = append(files, ReviewFile{
			path:    path,
			relPath: relPath,
			content: string(content),
			lines:   lines,
		})

		return nil
	})

	return files, err
}

// buildReviewPrompt creates the AI review prompt
func buildReviewPrompt(project *ProjectInfo, files []ReviewFile) string {
	var sb strings.Builder

	sb.WriteString("You are a senior code reviewer. Review the following project for bugs, security issues, performance problems, and bad logic patterns.\n\n")
	sb.WriteString(fmt.Sprintf("Project: %s (Type: %s)\n", project.Name, project.Type))
	sb.WriteString(fmt.Sprintf("Files: %d\n\n", len(files)))

	sb.WriteString("## Review Guidelines\n")
	sb.WriteString("- Look for: bugs, null/nil pointer dereferences, unhandled errors, SQL injection, XSS, race conditions, memory leaks, logic errors, missing validation\n")
	sb.WriteString("- Focus on HIGH and MEDIUM severity issues\n")
	sb.WriteString("- Be specific: reference exact file names and line numbers\n")
	sb.WriteString("- Do NOT suggest style/formatting changes unless they cause bugs\n")
	sb.WriteString("- Do NOT fix the code — only report findings\n\n")

	sb.WriteString("## Output Format\n")
	sb.WriteString("Respond in valid JSON only. No markdown, no explanation outside JSON.\n\n")
	sb.WriteString("```json\n")
	sb.WriteString("{\n")
	sb.WriteString("  \"findings\": [\n")
	sb.WriteString("    {\n")
	sb.WriteString("      \"file\": \"path/to/file.go\",\n")
	sb.WriteString("      \"line\": 42,\n")
	sb.WriteString("      \"severity\": \"HIGH|MEDIUM|LOW\",\n")
	sb.WriteString("      \"category\": \"Bug|Security|Performance|Logic|Style\",\n")
	sb.WriteString("      \"title\": \"Short title\",\n")
	sb.WriteString("      \"description\": \"Detailed explanation of the issue\",\n")
	sb.WriteString("      \"suggestion\": \"How to fix it\",\n")
	sb.WriteString("      \"code_snippet\": \"the problematic code\"\n")
	sb.WriteString("    }\n")
	sb.WriteString("  ],\n")
	sb.WriteString("  \"summary\": \"Overall assessment of the codebase quality\"\n")
	sb.WriteString("}\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## Code to Review\n\n")

	// Limit total code to ~15000 tokens to stay within model limits
	totalChars := 0
	maxChars := 45000 // ~15k tokens

	for _, f := range files {
		fileBlock := fmt.Sprintf("### File: %s (%d lines)\n```\n%s\n```\n\n", f.relPath, f.lines, f.content)
		if totalChars+len(fileBlock) > maxChars {
			sb.WriteString(fmt.Sprintf("### File: %s (%d lines)\n[TRUNCATED — file too large for full review]\n\n", f.relPath, f.lines))
			continue
		}
		sb.WriteString(fileBlock)
		totalChars += len(fileBlock)
	}

	return sb.String()
}

// performAIReview sends the code to an AI model for review
func performAIReview(config *ReviewConfig, prompt string) ([]ReviewFinding, string, error) {
	body := map[string]interface{}{
		"model": config.Model,
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are an expert code reviewer. You respond only in valid JSON. Never include markdown formatting or explanations outside the JSON object.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.2,
		"max_tokens":  4096,
	}

	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", config.BaseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Authorization", "Bearer "+config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		data, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(data))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, "", fmt.Errorf("response decode failed: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, "", fmt.Errorf("no response from AI model")
	}

	content := result.Choices[0].Message.Content

	// Parse the JSON response from the AI
	// Strip markdown code fences if present
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
	}
	content = strings.TrimSpace(content)

	var reviewResponse struct {
		Findings []ReviewFinding `json:"findings"`
		Summary  string          `json:"summary"`
	}

	if err := json.Unmarshal([]byte(content), &reviewResponse); err != nil {
		return nil, "", fmt.Errorf("AI response parse failed: %w\nRaw: %s", err, content[:min(200, len(content))])
	}

	return reviewResponse.Findings, reviewResponse.Summary, nil
}

// GenerateReviewReport creates the SHIPMATE_REVIEW.md file
func GenerateReviewReport(result *ReviewResult) error {
	var sb strings.Builder

	sb.WriteString("# 🚀 Shipmate Code Review Report\n\n")
	sb.WriteString(fmt.Sprintf("**Project:** %s  \n", result.ProjectName))
	sb.WriteString(fmt.Sprintf("**Type:** %s  \n", result.ProjectType))
	sb.WriteString(fmt.Sprintf("**Files Scanned:** %d  \n", result.FilesScanned))
	sb.WriteString(fmt.Sprintf("**Lines of Code:** %d  \n", result.TotalLines))
	sb.WriteString(fmt.Sprintf("**Reviewed At:** %s  \n", result.ReviewedAt))
	sb.WriteString(fmt.Sprintf("**Duration:** %s  \n\n", result.Duration))

	// Severity counts
	high, medium, low := 0, 0, 0
	for _, f := range result.Findings {
		switch strings.ToUpper(f.Severity) {
		case "HIGH":
			high++
		case "MEDIUM":
			medium++
		case "LOW":
			low++
		}
	}

	sb.WriteString("## Summary\n\n")
	if result.Summary != "" {
		sb.WriteString(result.Summary + "\n\n")
	}

	sb.WriteString(fmt.Sprintf("**Total Findings:** %d  \n", len(result.Findings)))
	sb.WriteString(fmt.Sprintf("| Severity | Count |\n|----------|-------|\n"))
	sb.WriteString(fmt.Sprintf("| 🔴 HIGH   | %d    |\n", high))
	sb.WriteString(fmt.Sprintf("| 🟡 MEDIUM | %d    |\n", medium))
	sb.WriteString(fmt.Sprintf("| 🔵 LOW    | %d    |\n\n", low))

	if len(result.Findings) == 0 {
		sb.WriteString("## ✅ No Issues Found\n\n")
		sb.WriteString("Great job! No significant issues were detected in your codebase.\n\n")
	} else {
		sb.WriteString("---\n\n## Findings\n\n")

		for i, f := range result.Findings {
			severity := "🔵"
			switch strings.ToUpper(f.Severity) {
			case "HIGH":
				severity = "🔴"
			case "MEDIUM":
				severity = "🟡"
			}

			sb.WriteString(fmt.Sprintf("### %d. %s %s — %s\n\n", i+1, severity, f.Severity, f.Title))
			sb.WriteString(fmt.Sprintf("**File:** `%s`", f.File))
			if f.Line > 0 {
				sb.WriteString(fmt.Sprintf(" (line %d)", f.Line))
			}
			sb.WriteString(fmt.Sprintf("  \n**Category:** %s\n\n", f.Category))
			sb.WriteString(fmt.Sprintf("**Description:**  \n%s\n\n", f.Description))

			if f.CodeSnippet != "" {
				sb.WriteString("**Code:**\n```")
				ext := filepath.Ext(f.File)
				if len(ext) > 1 {
					sb.WriteString(ext[1:])
				}
				sb.WriteString(fmt.Sprintf("\n%s\n```\n\n", f.CodeSnippet))
			}

			if f.Suggestion != "" {
				sb.WriteString(fmt.Sprintf("**Suggested Fix:**  \n%s\n\n", f.Suggestion))
			}

			sb.WriteString("---\n\n")
		}
	}

	sb.WriteString("*Generated by [Shipmate](https://myshipmate.cc) — The Smart Deployer for Developers*\n")

	return os.WriteFile("SHIPMATE_REVIEW.md", []byte(sb.String()), 0644)
}

// DisplayReviewSummary shows a terminal summary of the review
func DisplayReviewSummary(result *ReviewResult) {
	high, medium, low := 0, 0, 0
	for _, f := range result.Findings {
		switch strings.ToUpper(f.Severity) {
		case "HIGH":
			high++
		case "MEDIUM":
			medium++
		case "LOW":
			low++
		}
	}

	fmt.Println()
	if len(result.Findings) == 0 {
		fmt.Println("   ✅ No issues found! Your code looks clean.")
	} else {
		fmt.Printf("   Found %d issue(s):\n", len(result.Findings))
		if high > 0 {
			fmt.Printf("      🔴 HIGH:   %d\n", high)
		}
		if medium > 0 {
			fmt.Printf("      🟡 MEDIUM: %d\n", medium)
		}
		if low > 0 {
			fmt.Printf("      🔵 LOW:    %d\n", low)
		}
		fmt.Println()

		// Show top findings
		shown := 0
		for _, f := range result.Findings {
			if shown >= 5 {
				remaining := len(result.Findings) - 5
				if remaining > 0 {
					fmt.Printf("      ... and %d more (see SHIPMATE_REVIEW.md)\n", remaining)
				}
				break
			}

			severity := "🔵"
			switch strings.ToUpper(f.Severity) {
			case "HIGH":
				severity = "🔴"
			case "MEDIUM":
				severity = "🟡"
			}

			fmt.Printf("      %s %s — %s\n", severity, f.File, f.Title)
			shown++
		}
	}

	fmt.Println()
	fmt.Printf("   📄 Full report: SHIPMATE_REVIEW.md\n")
	fmt.Printf("   ⏱  Review took %s\n", result.Duration)
}
