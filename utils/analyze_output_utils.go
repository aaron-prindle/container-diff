package utils

type Result interface {
	OutputStruct() interface{}
	OutputText(resultType string) error
}

type AnalyzeResult interface {
	GetStruct() AnalyzeResult
	OutputText(analyzeType string) error
}

type ListAnalyzeResult struct {
	Image       string
	AnalyzeType string
	Analysis    []string
}

func (r ListAnalyzeResult) GetStruct() AnalyzeResult {
	return r
}

func (r ListAnalyzeResult) OutputText(analyzeType string) error {
	return TemplateOutput(r)
}

type MultiVersionPackageAnalyzeResult struct {
	Image       string
	AnalyzeType string
	Analysis    map[string]map[string]PackageInfo
}

func (r MultiVersionPackageAnalyzeResult) GetStruct() AnalyzeResult {
	return r
}

func (r MultiVersionPackageAnalyzeResult) OutputText(analyzeType string) error {
	return TemplateOutput(r)
}

type SingleVersionPackageAnalyzeResult struct {
	Image       string
	AnalyzeType string
	Analysis    map[string]PackageInfo
}

func (r SingleVersionPackageAnalyzeResult) GetStruct() AnalyzeResult {
	return r
}

func (r SingleVersionPackageAnalyzeResult) OutputText(diffType string) error {
	return TemplateOutput(r)
}

type FileAnalyzeResult struct {
	Image       string
	AnalyzeType string
	Analysis    []DirectoryEntry
}

func (r FileAnalyzeResult) GetStruct() AnalyzeResult {
	return r
}

func (r FileAnalyzeResult) OutputText(analyzeType string) error {
	return TemplateOutput(r)
}
