package plugin

// PluginValidateResponse is the JSON shape a plugin writes to stdout for "validate".
type PluginValidateResponse struct {
	Valid bool   `json:"valid"`
	Error string `json:"error,omitempty"`
}

// PluginArticle is the JSON shape a plugin writes per article for "fetch".
type PluginArticle struct {
	ExternalID  string `json:"external_id"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	Author      string `json:"author,omitempty"`
	Summary     string `json:"summary,omitempty"`
	Content     string `json:"content,omitempty"`
	PublishedAt string `json:"published_at,omitempty"` // RFC 3339
}

// PluginFetchResponse is the JSON shape a plugin writes to stdout for "fetch".
type PluginFetchResponse struct {
	Articles []PluginArticle `json:"articles,omitempty"`
	Error    string          `json:"error,omitempty"`
}
