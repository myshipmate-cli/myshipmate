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
	Category    string `json:"category"` // Bug, Security, Performance, Logic
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
	Usage        *ReviewUsage    `json:"usage,omitempty"`
}

// ReviewUsage shows plan and remaining reviews
type ReviewUsage struct {
	Plan      string `json:"plan"`
	Remaining int    `json:"remaining"`
}

// reviewAPIBase is the backend API URL
// Can be overridden with SHIPMATE_REVIEW_URL env var
func reviewAPIBase() string {
	if url := os.Getenv("SHIPMATE_REVIEW_URL"); url != "" {
		return url
	}
	return "https://myshipmate-review-api.fly.dev"
}

// ReviewCode performs a code review by calling the Shipmate Review API
func ReviewCode(project *ProjectInfo) (*ReviewResult, error) {
	start := time.Now()

	// Check authentication — requires Shipmate Cloud login
	token, err := GetToken("shipmate")
	if err != nil || token == "" {
		return nil, fmt.Errorf("code review requires a Shipmate Cloud account\n\n  Run: shipmate login shipmate\n  Then try again.\n\n  Free during beta — 5 reviews/month.\n  Visit: https://myshipmate.cc/pricing")
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
	fmt.Println("   🤖 Sending to Shipmate Review API...")

	// Build API request
	reqFiles := make([]map[string]interface{}, 0, len(files))
	for _, f := range files {
		reqFiles = append(reqFiles, map[string]interface{}{
			"path":    f.relPath,
			"content": f.content,
			"lines":   f.lines,
		})
	}

	reqBody := map[string]interface{}{
		"project_name": project.Name,
		"project_type": string(project.Type),
		"files":        reqFiles,
	}

	bodyBytes, _ := json.Marshal(reqBody)

	// Call the Shipmate Review API
	req, err := http.NewRequest("POST", reviewAPIBase()+"/api/v1/review", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("request creation failed: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	respData, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		// Parse error response
		var errResp map[string]interface{}
		json.Unmarshal(respData, &errResp)

		if resp.StatusCode == 401 {
			return nil, fmt.Errorf("authentication failed\n\n  Your Shipmate Cloud token is invalid or expired.\n  Run: shipmate login shipmate\n  Then try again.")
		}

		if resp.StatusCode == 402 {
			return nil, fmt.Errorf("free tier limit reached (5 reviews/month)\n\n  Upgrade to Shipmate Cloud Pro for unlimited reviews.\n  Visit: https://myshipmate.cc/pricing")
		}

		if errMsg, ok := errResp["error"].(string); ok {
			return nil, fmt.Errorf("review API error (%d): %s", resp.StatusCode, errMsg)
		}

		return nil, fmt.Errorf("review API error (%d): %s", resp.StatusCode, string(respData))
	}

	// Parse successful response
	var apiResp struct {
		Findings []ReviewFinding `json:"findings"`
		Summary  string          `json:"summary"`
		Usage    *ReviewUsage    `json:"usage"`
	}

	if err := json.Unmarshal(respData, &apiResp); err != nil {
		return nil, fmt.Errorf("response parse failed: %w", err)
	}

	elapsed := time.Since(start)

	result := &ReviewResult{
		ProjectName:  project.Name,
		ProjectType:  string(project.Type),
		FilesScanned: len(files),
		TotalLines:   totalLines,
		Findings:     apiResp.Findings,
		ReviewedAt:   time.Now().Format(time.RFC3339),
		Summary:      apiResp.Summary,
		Duration:     elapsed.Round(time.Millisecond).String(),
		Usage:        apiResp.Usage,
	}

	return result, nil
}

// ReviewFile represents a file being reviewed
type ReviewFile struct {
	relPath string
	content string
	lines   int
}

// collectReviewFiles gathers code files for review
func collectReviewFiles(project *ProjectInfo) ([]ReviewFile, error) {
	var files []ReviewFile

	// Extensions to review
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
			relPath: relPath,
			content: string(content),
			lines:   lines,
		})

		return nil
	})

	return files, err
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

	if result.Usage != nil {
		sb.WriteString(fmt.Sprintf("**Plan:** %s | **Reviews remaining:** %d this month\n\n", result.Usage.Plan, result.Usage.Remaining))
	}

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

	if result.Usage != nil {
		fmt.Println()
		fmt.Printf("   📊 Plan: %s | %d reviews remaining this month\n", result.Usage.Plan, result.Usage.Remaining)
	}

	fmt.Println()
	fmt.Printf("   📄 Full report: SHIPMATE_REVIEW.md\n")
	fmt.Printf("   ⏱  Review took %s\n", result.Duration)
}
