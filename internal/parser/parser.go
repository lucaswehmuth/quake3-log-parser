package parser

// Parser defines the interface for parsing logs
type Parser interface {
	FetchLogFileFromUrl(url string) (string, error)
	ParseLogContent(lines []string) interface{}
	PrintReport(parsedData interface{})
}
