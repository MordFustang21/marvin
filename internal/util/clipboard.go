package util

import (
	"fmt"
	"os/exec"
	"runtime"
)

// CopyToClipboard copies the given text to the system clipboard
func CopyToClipboard(text string) error {
	switch runtime.GOOS {
	case "darwin":
		// macOS uses pbcopy command
		cmd := exec.Command("pbcopy")
		pipe, err := cmd.StdinPipe()
		if err != nil {
			return fmt.Errorf("failed to create stdin pipe: %w", err)
		}
		
		// Start the command before sending input
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start pbcopy: %w", err)
		}
		
		// Write the text to the pipe
		if _, err := pipe.Write([]byte(text)); err != nil {
			return fmt.Errorf("failed to write to pipe: %w", err)
		}
		
		// Close the pipe
		if err := pipe.Close(); err != nil {
			return fmt.Errorf("failed to close pipe: %w", err)
		}
		
		// Wait for the command to finish
		return cmd.Wait()
		
	case "linux":
		// For Linux we could use xsel, xclip, or wl-copy depending on the environment
		// This is a simplified approach that tries xclip first
		cmd := exec.Command("xclip", "-selection", "clipboard")
		pipe, err := cmd.StdinPipe()
		if err != nil {
			return fmt.Errorf("failed to create stdin pipe: %w", err)
		}
		
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start xclip, clipboard functionality may not be available: %w", err)
		}
		
		if _, err := pipe.Write([]byte(text)); err != nil {
			return fmt.Errorf("failed to write to pipe: %w", err)
		}
		
		if err := pipe.Close(); err != nil {
			return fmt.Errorf("failed to close pipe: %w", err)
		}
		
		return cmd.Wait()
		
	case "windows":
		// For Windows, using PowerShell to set clipboard
		cmd := exec.Command("powershell", "-command", "Set-Clipboard", "-Value", text)
		return cmd.Run()
		
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// GetFromClipboard retrieves text from the system clipboard
func GetFromClipboard() (string, error) {
	var cmd *exec.Cmd
	
	switch runtime.GOOS {
	case "darwin":
		// macOS uses pbpaste command
		cmd = exec.Command("pbpaste")
		
	case "linux":
		// For Linux, try xclip
		cmd = exec.Command("xclip", "-selection", "clipboard", "-o")
		
	case "windows":
		// For Windows, using PowerShell to get clipboard
		cmd = exec.Command("powershell", "-command", "Get-Clipboard")
		
	default:
		return "", fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get clipboard content: %w", err)
	}
	
	return string(output), nil
}