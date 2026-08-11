package grounding

type Metadata struct {
	SearchQueries []string `json:"search_queries"`
	Sources       []Source `json:"sources"`
}

type Source struct {
	Title      string `json:"title"`
	Publisher  string `json:"publisher,omitempty"`
	URL        string `json:"url"`
	ObservedAt string `json:"observed_at,omitempty"`
}
