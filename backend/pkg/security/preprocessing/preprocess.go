package preprocessing

type ProcessedText struct {
	Text           string
	OriginalLength int
}

type Preprocessor interface {
	Process(text string) ProcessedText
}

type DefaultPreprocessor struct{}

func (p *DefaultPreprocessor) Process(text string) ProcessedText {
	return ProcessedText{Text: text, OriginalLength: len(text)}
}
