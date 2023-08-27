package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	var dialogBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		MaxWidth(85).
		BorderTop(true).
		BorderLeft(true).
		BorderRight(true).
		BorderBottom(true)

	var dialogLineStyle = lipgloss.NewStyle().
		// Align(lipgloss.Left).
		Width(77)

	var headerStyle = lipgloss.NewStyle().
		Inherit(dialogLineStyle).
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4"))

	var redTextStyle = lipgloss.NewStyle().
		Inherit(dialogLineStyle).
		Bold(true).
		Foreground(lipgloss.Color("#FF0000"))

	var greenTextStyle = lipgloss.NewStyle().
		Inherit(dialogLineStyle).
		Bold(true).
		Foreground(lipgloss.Color("#00FF00"))

	// str := "Description: (Unfinished) One man's junk is another man's treasure. This 102934rj90sad90sdjv aosdfjaisdofj  ajo sjo d"
	// str = dialogLineStyle.Render(str)
	// println(dialogBoxStyle.Render(str))

	githubToken, githubTokenSet := os.LookupEnv("GITHUB_API_TOKEN")

	if len(os.Args) != 2 {
		println("Usage: github-profile-scanner.exe [GitHub Username]")
		return
	}

	username := os.Args[1]
	println("Scanning GitHub profile of " + username + "...")

	var client *github.Client

	ctx := context.Background()
	if githubTokenSet {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: githubToken},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
		println("Using GitHub token")
	} else {
		c := http.Client{Timeout: time.Duration(1) * time.Second}
		client = github.NewClient(&c)
		println("Using unauthenticated client")
	}

	// list all repositories for the authenticated user
	repos, _, err := client.Repositories.List(ctx, username, nil)

	if err != nil {
		println("Error: " + err.Error())
		return
	}

	missingDescriptionCount := 0
	missiongReadmeCount := 0
	missingImageCount := 0

	for _, repo := range repos {
		if *repo.Fork {
			continue
		}

		r := strings.Builder{}
		r.WriteString(headerStyle.Render(*repo.Name))
		r.WriteRune('\n')

		description := repo.GetDescription()
		if description == "" {
			r.WriteString(redTextStyle.Render(fmt.Sprintf("✕ No description for %s", *repo.Name)))
			r.WriteRune('\n')
			missingDescriptionCount++
		} else {
			r.WriteString(greenTextStyle.Render(fmt.Sprintf("✓ Description: %s", description)))
			r.WriteRune('\n')
		}

		// defaultBranch := repo.GetDefaultBranch()
		// println(defaultBranch)

		readmeContent, _, _, err := client.Repositories.GetContents(ctx, username, *repo.Name, "README.md", nil)
		if err != nil {
			r.WriteString(redTextStyle.Render(fmt.Sprintf("✕ No README.md for %s", *repo.Name)))
			r.WriteRune('\n')
			missiongReadmeCount++
			missingImageCount++
		} else {
			r.WriteString(greenTextStyle.Render("✓ README.md"))
			r.WriteRune('\n')

			readmeContentText, err := readmeContent.GetContent()

			if err != nil {
				r.WriteString(redTextStyle.Render("Error getting readme content: " + err.Error()))
				missingImageCount++
			} else if strings.Contains(readmeContentText, "![") {
				r.WriteString(greenTextStyle.Render("✓ Image in README.md"))
				r.WriteRune('\n')
			} else {
				r.WriteString(redTextStyle.Render("✕ No image in README.md"))
				r.WriteRune('\n')
				missingImageCount++
			}

		}

		println(dialogBoxStyle.Render(r.String()))

		println()
	}

	if missingDescriptionCount > 0 {
		println(redTextStyle.Render(fmt.Sprintf("%d repositories missing description", missingDescriptionCount)))
	} else {
		println(greenTextStyle.Render("All repositories have descriptions"))
	}

	if missiongReadmeCount > 0 {
		println(redTextStyle.Render(fmt.Sprintf("%d repositories missing README.md", missiongReadmeCount)))
	} else {
		println(greenTextStyle.Render("All repositories have README.md"))
	}

	if missingImageCount > 0 {
		println(redTextStyle.Render(fmt.Sprintf("%d repositories missing images", missingImageCount)))
	} else {
		println(greenTextStyle.Render("All repositories have images"))
	}
	fmt.Printf("of %d repositories\n", len(repos))

}
