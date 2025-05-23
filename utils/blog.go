package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// ExtractKeywordsFromContent automatically extracts keywords from blog content
func ExtractKeywordsFromContent(title, content string) string {
	// Combine title and content for keyword extraction
	text := strings.ToLower(title + " " + content)

	// Remove HTML tags if any
	re := regexp.MustCompile(`<[^>]*>`)
	text = re.ReplaceAllString(text, " ")

	// Split into words
	words := strings.FieldsFunc(text, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})

	// Filter out common stop words and short words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "up": true, "about": true, "into": true,
		"through": true, "during": true, "before": true, "after": true, "above": true,
		"below": true, "between": true, "among": true, "is": true, "are": true, "was": true,
		"were": true, "be": true, "been": true, "being": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "must": true, "can": true,
		"this": true, "that": true, "these": true, "those": true, "i": true, "me": true,
		"my": true, "myself": true, "we": true, "our": true, "ours": true, "ourselves": true,
		"you": true, "your": true, "yours": true, "yourself": true, "yourselves": true,
		"he": true, "him": true, "his": true, "himself": true, "she": true, "her": true,
		"hers": true, "herself": true, "it": true, "its": true, "itself": true, "they": true,
		"them": true, "their": true, "theirs": true, "themselves": true,
	}

	// Count word frequency
	wordCount := make(map[string]int)
	for _, word := range words {
		if len(word) >= 3 && !stopWords[word] {
			wordCount[word]++
		}
	}

	// Get most frequent words (max 10)
	var keywords []string
	maxKeywords := 10
	minFreq := 1

	if len(wordCount) > maxKeywords {
		minFreq = 2 // Increase minimum frequency if too many words
	}

	for word, count := range wordCount {
		if count >= minFreq && len(keywords) < maxKeywords {
			keywords = append(keywords, word)
		}
	}

	return strings.Join(keywords, ",")
}

// ValidateImageURL checks if a URL appears to be a valid image URL
func ValidateImageURL(url string) bool {
	if url == "" {
		return true // Empty URL is allowed
	}

	// Basic URL validation
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return false
	}

	// Check for common image extensions
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".bmp"}
	lowerURL := strings.ToLower(url)

	for _, ext := range imageExtensions {
		if strings.Contains(lowerURL, ext) {
			return true
		}
	}

	// If no extension found, might still be valid (e.g., dynamic image URLs)
	return true
}

// GenerateSlug creates a URL-friendly slug from title
func GenerateSlug(title string) string {
	// Convert to lowercase
	slug := strings.ToLower(title)

	// Replace spaces and special characters with hyphens
	re := regexp.MustCompile(`[^a-z0-9\s-]`)
	slug = re.ReplaceAllString(slug, "")

	// Replace multiple spaces/hyphens with single hyphen
	re = regexp.MustCompile(`[\s-]+`)
	slug = re.ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	// Limit length
	if len(slug) > 100 {
		slug = slug[:100]
		slug = strings.Trim(slug, "-")
	}

	return slug
}

// TruncateText truncates text to specified length with ellipsis
func TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}

	truncated := text[:maxLength]

	// Try to break at word boundary
	lastSpace := strings.LastIndex(truncated, " ")
	if lastSpace > maxLength/2 {
		truncated = truncated[:lastSpace]
	}

	return truncated + "..."
}

// StripHTMLTags removes HTML tags from text
func StripHTMLTags(text string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(text, "")
}

// CalculateReadingTime estimates reading time based on word count
func CalculateReadingTime(content string) int {
	// Average reading speed: 200 words per minute
	words := strings.Fields(StripHTMLTags(content))
	wordCount := len(words)
	readingTime := wordCount / 200

	if readingTime < 1 {
		return 1
	}

	return readingTime
}
