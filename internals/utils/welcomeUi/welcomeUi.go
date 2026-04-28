package welcomeUi

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

func ShowWelcomeUI() {
	// Palette
	mint := lipgloss.Color("#9FDBA1")
	muted := lipgloss.Color("#eee")

	// ===== Banner =====
	banner := `
   ██████╗  ██████╗  ██████╗ ██╗████████╗ ██████╗
  ██╔════╝ ██╔═══██╗██╔════╝ ██║╚══██╔══╝██╔═══██╗
  ██║      ██║   ██║██║  ███╗██║   ██║   ██║   ██║
  ██║      ██║   ██║██║   ██║██║   ██║   ██║   ██║
  ╚██████╗ ╚██████╔╝╚██████╔╝██║   ██║   ╚██████╔╝
   ╚═════╝  ╚═════╝  ╚═════╝ ╚═╝   ╚═╝    ╚═════╝
	`

	styledBanner := lipgloss.NewStyle().
		Foreground(mint).
		Bold(true).
		Render(banner)

	// Subtitle
	subtitle := lipgloss.NewStyle().
		Foreground(muted).
		Render("Token Optimizer for AI CLI")

	// Command arrow
	arrow := lipgloss.NewStyle().
		Foreground(mint).
		Render("▸")

	cmd := func(label, desc string) string {
		return fmt.Sprintf(
			"%s %-22s %s",
			arrow,
			label,
			lipgloss.NewStyle().
				Foreground(muted).
				Render(desc),
		)
	}

	// Main commands
	commands := lipgloss.JoinVertical(
		lipgloss.Left,
		cmd("cogito install", "setup and configure Cogito"),
		cmd("cogito build-map", "generate codebase substrate map"),
		cmd("cogito uninstall", "remove Cogito and cleanup"),
		cmd("cogito --help", "show help information"),
		cmd("cogito -v", "show current version"),
	)

	// Internal/testing commands
	internal := lipgloss.JoinVertical(
		lipgloss.Left,
		cmd("cogito serve-mcp", "internal MCP stdio server"),
	)

	divider := lipgloss.NewStyle().
		Foreground(muted).
		Render("────────────────────────────────────────")

	note := lipgloss.NewStyle().
		Foreground(muted).
		Render("serve-mcp is for Codex/Claude MCP integration, not normal terminal use")

	// Final box
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(muted).
		Padding(1, 3).
		Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				styledBanner,
				"",
				subtitle,
				"",
				divider,
				"",
				commands,
				"",
				divider,
				"Internal / Testing",
				internal,
				"",
				note,
			),
		)

	fmt.Println(box)
}
