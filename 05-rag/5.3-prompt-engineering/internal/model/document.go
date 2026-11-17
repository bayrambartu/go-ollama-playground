package model

type Document struct {
	ID      int64  `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CreateDocumentRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type SearchRequest struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k"`
}

type SearchResult struct {
	Document
	Similarity float64 `json:"similarity"`
}
