package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sort"
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
		Width(77)

	var headerStyle = lipgloss.NewStyle().
		Inherit(dialogLineStyle).
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4"))
	var headerTitleStyle = lipgloss.NewStyle().
		Padding(0, 0, 0, 1).
		Width(50)
	var headerDateStyle = lipgloss.NewStyle().
		Width(27).
		Padding(0, 1, 0, 0).
		Align(lipgloss.Right)

	var redTextStyle = lipgloss.NewStyle().
		Inherit(dialogLineStyle).
		Bold(true).
		Foreground(lipgloss.Color("#FF0000"))

	var greenTextStyle = lipgloss.NewStyle().
		Inherit(dialogLineStyle).
		Bold(true).
		Foreground(lipgloss.Color("#00FF00"))

	var orangeTextStyle = lipgloss.NewStyle().
		Inherit(dialogLineStyle).
		Bold(true).
		Foreground(lipgloss.Color("#FFA500"))

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
	missingReadmeCount := 0
	shortReadmeCount := 0
	missingImageCount := 0
	masterBranchCount := 0

	sort.SliceStable(repos, func(i, j int) bool {
		return repos[i].GetUpdatedAt().Before(repos[j].GetUpdatedAt().Time)
	})

	for _, repo := range repos {
		if *repo.Fork {
			continue
		}

		r := strings.Builder{}
		headerTextRendering := headerTitleStyle.Render(fmt.Sprintf("%s (%s)", *repo.Name, *repo.Language))
		headerDateRendering := headerDateStyle.Render("Last updated", repo.GetUpdatedAt().Format("2006-01-02"))
		r.WriteString(headerStyle.Render(headerTextRendering + headerDateRendering))

		description := repo.GetDescription()
		if description == "" {
			r.WriteRune('\n')
			r.WriteString(redTextStyle.Render("✕ No description"))

			missingDescriptionCount++
		} else {
			r.WriteRune('\n')
			r.WriteString(greenTextStyle.Render(fmt.Sprintf("✓ Description: %s", description)))
		}

		defaultBranch := repo.GetDefaultBranch()
		if defaultBranch == "master" {
			r.WriteRune('\n')
			r.WriteString(orangeTextStyle.Render("✕ Default branch is master"))
			masterBranchCount++
		} else {
			r.WriteRune('\n')
			r.WriteString(greenTextStyle.Render(fmt.Sprintf("✓ Default branch is %s", defaultBranch)))
		}

		readmeContent, _, _, err := client.Repositories.GetContents(ctx, username, *repo.Name, "README.md", nil)
		if err != nil {
			r.WriteRune('\n')
			r.WriteString(redTextStyle.Render("✕ No README.md"))

			missingReadmeCount++
			missingImageCount++
		} else {
			r.WriteRune('\n')
			r.WriteString(greenTextStyle.Render("✓ README.md"))

			readmeContentText, err := readmeContent.GetContent()

			if err != nil {
				r.WriteString(redTextStyle.Render("Error getting readme content: " + err.Error()))
				missingImageCount++
				shortReadmeCount++
			} else {
				if strings.Contains(readmeContentText, "![") {
					r.WriteRune('\n')
					r.WriteString(greenTextStyle.Render("✓ Image in README.md"))
				} else {
					r.WriteRune('\n')
					r.WriteString(redTextStyle.Render("✕ No image in README.md"))
					missingImageCount++
				}

				if len(readmeContentText) < 100 {
					r.WriteRune('\n')
					r.WriteString(redTextStyle.Render(fmt.Sprintf("✕ README.md is too short (%d characters)", len(readmeContentText))))
					shortReadmeCount++
				} else {
					r.WriteRune('\n')
					r.WriteString(greenTextStyle.Render(fmt.Sprintf("✓ README.md is long enough (%d characters)", len(readmeContentText))))
				}
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

	if missingReadmeCount > 0 {
		println(redTextStyle.Render(fmt.Sprintf("%d repositories missing README.md", missingReadmeCount)))
	} else {
		println(greenTextStyle.Render("All repositories have README.md"))
	}

	if missingImageCount > 0 {
		println(redTextStyle.Render(fmt.Sprintf("%d repositories missing images", missingImageCount)))
	} else {
		println(greenTextStyle.Render("All repositories have images"))
	}

	if shortReadmeCount > 0 {
		println(redTextStyle.Render(fmt.Sprintf("%d repositories have a short README.md", shortReadmeCount)))
	} else {
		println(greenTextStyle.Render("All repositories have a long-enough README.md"))
	}

	if masterBranchCount > 0 {
		println(orangeTextStyle.Render(fmt.Sprintf("%d repositories have master as default branch", masterBranchCount)))
	} else {
		println(greenTextStyle.Render("All repositories have non-master default branch"))
	}

	fmt.Printf("of %d repositories\n", len(repos))

}
