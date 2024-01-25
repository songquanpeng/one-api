package aiproxy

type LibraryRequest struct {
	Model     string `json:"model"`
	Query     string `json:"query"`
	LibraryId string `json:"libraryId"`
	Stream    bool   `json:"stream"`
}

type LibraryError struct {
	ErrCode int    `json:"errCode"`
	Message string `json:"message"`
}

type LibraryDocument struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type LibraryResponse struct {
	Success   bool              `json:"success"`
	Answer    string            `json:"answer"`
	Documents []LibraryDocument `json:"documents"`
	LibraryError
}

type LibraryStreamResponse struct {
	Content   string            `json:"content"`
	Finish    bool              `json:"finish"`
	Model     string            `json:"model"`
	Documents []LibraryDocument `json:"documents"`
}
