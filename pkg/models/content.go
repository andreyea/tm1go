package models

// Content represents a TM1 content item (file or folder) with optional nested contents.
type Content struct {
	OdataType        string    `json:"@odata.type,omitempty"`
	ID               string    `json:"ID,omitempty"`
	Name             string    `json:"Name,omitempty"`
	Size             int       `json:"Size,omitempty"`
	LastUpdated      string    `json:"LastUpdated,omitempty"`
	MediaContentType string    `json:"Content@odata.mediaContentType,omitempty"`
	Contents         []Content `json:"Contents,omitempty"`
}

func extractNamesFromContents(contents []Content, parentPath string) []string {
	var names []string

	for _, content := range contents {
		currentPath := content.Name
		if parentPath != "" {
			currentPath = parentPath + "/" + content.Name
		}

		names = append(names, currentPath)

		if len(content.Contents) > 0 {
			names = append(names, extractNamesFromContents(content.Contents, currentPath)...)
		}
	}

	return names
}

// ExtractNamesFromContents returns a flat list of paths for the provided contents.
func ExtractNamesFromContents(contents []Content) []string {
	return extractNamesFromContents(contents, "")
}

func extractNamesFromContent(content Content, parentPath string) []string {
	var names []string

	currentPath := content.Name
	if parentPath != "" {
		currentPath = parentPath + "/" + content.Name
	}
	names = append(names, currentPath)

	for _, nestedContent := range content.Contents {
		names = append(names, extractNamesFromContent(nestedContent, currentPath)...)
	}

	return names
}

// ExtractNamesFromContent returns a flat list of paths for the provided content item.
func ExtractNamesFromContent(content Content) []string {
	return extractNamesFromContent(content, "")
}
