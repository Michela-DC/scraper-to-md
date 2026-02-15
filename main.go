package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
)

const ethicalSleepTime = 500 * time.Millisecond

type Creature struct {
	Title       string
	Subtitle    string // Challenge Rating
	ImageURL    string
	AC          string
	HP          string
	Speed       string
	Stats       map[string]string
	Skills      string
	Senses      string
	Languages   string
	Proficiency string
	Actions     []Action
	Description string
	Type        string
	Extra       map[string]string
}

type Action struct {
	Name        string
	Description string
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "scrape [url]",
		Short: "Scrapes D&D beasts and generates a markdown file",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			scrapeMainPage(args[0])
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func scrapeMainPage(url string) {
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	outputFile, _ := os.Create("out/beasts.md")
	defer outputFile.Close()

	doc.Find("div[style*='width: 33%'] p a:not(.newpage):not([href*='#toc'])").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		subPath := "http://dndroll.wikidot.com" + href
		fmt.Printf("Scraping: %s\n", subPath)

		creature := scrapeCreaturePage(subPath)
		writeMarkdown(outputFile, creature)
		time.Sleep(ethicalSleepTime) // avoid DDoS
	})
}

func scrapeCreaturePage(url string) Creature {
	res, err := http.Get(url)
	if err != nil {
		return Creature{}
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	c := Creature{Stats: make(map[string]string), Extra: make(map[string]string)}

	c.Title = strings.TrimSpace(doc.Find("#page-title").Text())
	content := doc.Find("#page-content")

	imgTag := content.Find("img.image")
	if imgTag.Length() > 0 {
		src, exists := imgTag.Attr("src")
		if exists {
			if !strings.HasPrefix(src, "http") {
				c.ImageURL = "http://dndroll.wikidot.com" + src
			} else {
				c.ImageURL = src
			}
		}
	}

	rawHTML, _ := content.Html()
	normalizedHTML := strings.NewReplacer("<br>", "\n", "<br/>", "\n", "<BR>", "\n").Replace(rawHTML)

	// Use a temporary document to strip remaining HTML tags while preserving our newlines
	tempDoc, _ := goquery.NewDocumentFromReader(strings.NewReader(normalizedHTML))
	lines := strings.Split(tempDoc.Text(), "\n")
	isDescription := false
	isActions := false
	for _, rawLine := range lines {
		text := strings.TrimSpace(rawLine)
		if text == "" {
			continue
		}
		lowerText := strings.ToLower(text)
		if lowerText == "actions" {
			isActions = true
			isDescription = false
			continue
		}
		if lowerText == "description" {
			isDescription = true
			isActions = false
			continue
		}
		if isDescription {
			c.Description += text + "\n\n"
			continue
		}

		switch {
		case strings.Contains(text, "Armor Class"):
			c.AC = strings.TrimSpace(strings.TrimPrefix(text, "Armor Class"))
		case strings.Contains(text, "Hit Points"):
			c.HP = strings.TrimSpace(strings.TrimPrefix(text, "Hit Points"))
		case strings.Contains(text, "Speed"):
			c.Speed = extractSpeed(text)
		case strings.HasPrefix(text, "Challenge"):
			c.Subtitle = strings.Fields(text)[1]
		case strings.HasPrefix(text, "Skills"):
			c.Skills = strings.TrimSpace(strings.TrimPrefix(text, "Skills"))
		case strings.HasPrefix(text, "Senses"):
			c.Senses = strings.TrimSpace(strings.TrimPrefix(text, "Senses"))
		case strings.HasPrefix(text, "Languages"):
			c.Languages = strings.TrimSpace(strings.TrimPrefix(text, "Languages"))
		case strings.HasPrefix(text, "Proficiency Bonus"):
			c.Proficiency = strings.TrimSpace(strings.TrimPrefix(text, "Proficiency Bonus"))
		case isActions || strings.Contains(text, "Attack:"):
			name := strings.Split(text, ".")[0]
			c.Actions = append(c.Actions, Action{
				Name:        name,
				Description: strings.TrimSpace(strings.TrimPrefix(text, name+".")),
			})
		case strings.Contains(text, "beast"):
			match, _ := regexp.MatchString(`\w+ beast, \w+`, text)
			if match {
				c.Type = text
				continue
			}
			fallthrough

		default:
			// Capture Traits (Keen Smell, etc.) or "Extra" info
			// Only if it looks like "Name. Description"
			if !strings.Contains(text, ".") {
				continue
			}
			parts := strings.SplitN(text, ".", 2)
			name := strings.TrimSpace(parts[0])

			// Filter out source citations
			if name != "Source" && !isDescription {
				c.Extra[name] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Ability Scores Table
	doc.Find("table.wiki-content-table tr").Each(func(i int, s *goquery.Selection) {
		cells := s.Find("td")
		if cells.Length() < 3 {
			return
		}

		name := strings.TrimSpace(cells.Eq(0).Text())
		score := strings.TrimSpace(cells.Eq(1).Text())
		bonus := strings.TrimSpace(cells.Eq(2).Text())

		var key string
		switch name {
		case "Strength":
			key = "STR"
		case "Dexterity":
			key = "DEX"
		case "Constitution":
			key = "CON"
		case "Intelligence":
			key = "INT"
		case "Wisdom":
			key = "WIS"
		case "Charisma":
			key = "CHA"
		}
		if key != "" {
			c.Stats[key] = fmt.Sprintf("%s (%s)", score, bonus)
		}
	})

	return c
}

func extractSpeed(text string) string {
	rawSpeed := strings.TrimSpace(strings.TrimPrefix(text, "Speed"))
	re := regexp.MustCompile(`(\d+)\s*ft`)
	match := re.FindStringSubmatch(rawSpeed)

	if len(match) <= 1 {
		return rawSpeed
	}

	// match[1] is the first capture group (the number)
	speedValue, err := strconv.Atoi(match[1])
	if err != nil {
		return rawSpeed
	}
	var speed string
	speedSq := float64(speedValue) / 5.0
	speedMt := speedSq * 1.5

	// Format: "20 ft (4 q, 6 m), burrow 5 ft."
	// If there are other speeds (burrow, fly), append them
	speed = fmt.Sprintf("%d ft (%.0f q, %.0f m)", speedValue, speedSq, speedMt)
	if strings.Contains(rawSpeed, ",") {
		extraSpeeds := strings.SplitN(rawSpeed, ",", 2)
		if len(extraSpeeds) > 1 {
			speed += "," + extraSpeeds[1]
		}
	}
	return speed
}

func printIfPresent(f *os.File, prefix, content string) {
	if content != "" {
		fmt.Fprintf(f, "%s %s\n", prefix, content)
	}
}

func writeMarkdown(f *os.File, c Creature) {
	fmt.Fprintln(f, "{{monster,frame,wide")

	fmt.Fprintln(f, "{{wide")
	fmt.Fprintf(f, "# %s\n", c.Title)
	fmt.Fprintf(f, "### Challenge Rating: %s\n", c.Subtitle)
	fmt.Fprintf(f, "##### %s\n\n", c.Type)
	fmt.Fprintln(f, "}}")
	fmt.Fprintln(f, "")

	// Placeholder image
	if c.ImageURL == "" {
		c.ImageURL = "https://www.pngfind.com/pngs/m/266-2663967_d-d-logo-png-dungeons-dragons-transparent-png.png"
	}
	fmt.Fprintf(f, "<img src=\"%s\" width=\"300\" />\n\n", c.ImageURL)

	fmt.Fprintln(f, "::::")

	fmt.Fprintln(f, "___")
	fmt.Fprintf(f, "- **Armor Class:** :: %s\n", c.AC)
	fmt.Fprintf(f, "- **Hit Points:** :: %s\n", c.HP)
	fmt.Fprintf(f, "- **Speed:** :: %s\n", c.Speed)
	fmt.Fprintln(f, "___")
	printIfPresent(f, "- **Skills:** ::", c.Skills)
	printIfPresent(f, "- **Senses:** ::", c.Senses)
	printIfPresent(f, "- **Languages:** ::", c.Languages)
	printIfPresent(f, "- **Proficiency Bonus:** ::", c.Proficiency)
	fmt.Fprintln(f, "___")

	fmt.Fprintln(f, "| STR | DEX | CON | INT | WIS | CHA |")
	fmt.Fprintln(f, "|:---:|:---:|:---:|:---:|:---:|:---:|")
	fmt.Fprintf(f, "| %s | %s | %s | %s | %s | %s |\n\n",
		c.Stats["STR"], c.Stats["DEX"], c.Stats["CON"], c.Stats["INT"], c.Stats["WIS"], c.Stats["CHA"])
	fmt.Fprintln(f, "___")

	fmt.Fprintln(f, "### Actions")
	for _, a := range c.Actions {
		fmt.Fprintf(f, "- **%s:** :: %s\n", a.Name, a.Description)
	}

	fmt.Fprintln(f, "")
	fmt.Fprintln(f, "::::")
	fmt.Fprintln(f, "")

	fmt.Fprintf(f, "{{descriptive,wide")
	if c.Description != "" {
		fmt.Fprintf(f, "\n### Description\n%s\n\n---\n\n", c.Description)
	}

	if len(c.Extra) != 0 {
		fmt.Fprintln(f, "\n### Extra")
	}
	for k, v := range c.Extra {
		fmt.Fprintf(f, "- **%s:** %s\n", k, v)
	}

	fmt.Fprintln(f, "}}")

	fmt.Fprintln(f, "}}")
	fmt.Fprintln(f, "\\page")
}
