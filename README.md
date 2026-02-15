# DnD beasts scraper
Scrape a website and create a md file containing DnD creatures.

## Context
As a D&D player, I often found myself needing a way to reference creature stats quickly without relying on a stable internet connection or navigating slow web interfaces during a session. 

I decided to turn this necessity into a learning opportunity. This project serves as a deep dive into the **Go CLI ecosystem**, specifically focusing on:
* **Structured Data Extraction:** Handling inconsistent and "messy" HTML patterns (nested tables, raw line breaks, and mixed formatting).
* **CLI UX:** Building a robust command-line interface using `Cobra`.
* **Data Transformation:** Converting unstructured web content into clean, portable Markdown files.
* **Speed Conversion:** Automatically calculates grid squares (q) and meters (m) from standard feet measurements using Regex.
* **Image Embedding:** Automatically finds and links creature artwork within the generated Markdown.
* **Markdown Output:** Generates clean, readable `.md` files compatible with the [Homebrewery](https://homebrewery.naturalcrit.com/).


## Prerequisites
- [Go 1.24+](https://go.dev/dl/)
- Internet connection

## Usage
```shell
Scrapes D&D beasts and generates a markdown file

Usage:
  scrape [url] [flags]

Flags:
  -h, --help   help for scrape
```

From CLI run 

```shell
go run . [url]
```


## Roadmap
- Set output file directory
- Better page parsing
- Leverage cobra for flags parsing and validation
- Placeholder image as CLI arg
- Fallback image search
- MD format (currently hardcoded to homebrewery)
- Choose output format (pdf, md, html, sql, etc)
- More features?

## Disclaimer & Ethical Use
This repository is a personal learning project created for the purpose of exploring the Go programming language, the Cobra CLI framework, and the GoQuery library.

### Educational Purpose
The code provided here is for educational demonstrations only. It was developed to practice parsing complex, non-standard HTML structures and implementing state-machine logic in web scrapers.

### Scraping Ethics
Users of this tool should be aware of the following:

**Terms of Service**: Always check the website's robots.txt and Terms of Service before scraping.

**Respect the Servers**: This tool is intended for single-page lookups. Do not use it to perform high-frequency requests that could degrade the target site's performance.

**Copyright**: The content being scraped belongs to the original creators and publishers. This tool is not intended for commercial redistribution of copyrighted data.

**Personal Use**: This project should only be used for personal, offline data organization.