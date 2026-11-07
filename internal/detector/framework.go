package detector

import (
	"os"
	"path/filepath"
	"strings"
)

// DetectFramework inspects dependency files to determine the application framework.
func DetectFramework(dir, language string) string {
	switch language {
	case "go":
		content, _ := os.ReadFile(filepath.Join(dir, "go.mod"))
		text := string(content)
		if strings.Contains(text, "github.com/gin-gonic/gin") {
			return "gin"
		}
		if strings.Contains(text, "github.com/gofiber/fiber") {
			return "fiber"
		}
		if strings.Contains(text, "github.com/labstack/echo") {
			return "echo"
		}
		if strings.Contains(text, "net/http") || strings.Contains(text, "github.com/spf13/cobra") {
			return "standard"
		}

	case "node":
		content, _ := os.ReadFile(filepath.Join(dir, "package.json"))
		text := string(content)
		if strings.Contains(text, `"next"`) {
			return "nextjs"
		}
		if strings.Contains(text, `"express"`) {
			return "express"
		}
		if strings.Contains(text, `"@nestjs/core"`) {
			return "nest"
		}
		if strings.Contains(text, `"react"`) {
			return "react"
		}

	case "python":
		var text string
		if data, err := os.ReadFile(filepath.Join(dir, "requirements.txt")); err == nil {
			text = string(data)
		} else if data, err := os.ReadFile(filepath.Join(dir, "pyproject.toml")); err == nil {
			text = string(data)
		}
		if strings.Contains(text, "fastapi") {
			return "fastapi"
		}
		if strings.Contains(text, "django") {
			return "django"
		}
		if strings.Contains(text, "flask") {
			return "flask"
		}
	}

	return ""
}
