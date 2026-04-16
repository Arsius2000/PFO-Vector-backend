package model




type RowError struct{
	Row    int    `json:"row"`
    Field  string `json:"field"`
    Reason string `json:"reason"`
}
type ImportResult struct {
    TotalRows int        `json:"total_rows"`
    Created   int        `json:"created"`
    Failed    int        `json:"failed"`

    Errors    []RowError `json:"errors"`
}