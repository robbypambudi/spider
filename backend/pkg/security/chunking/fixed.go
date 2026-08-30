package chunking

type Chunk struct {
	Index    int
	Text     string
	Start    *int
	End      *int
	Metadata map[string]interface{}
}

type Chunker interface {
	Name() string
	Chunk(text string) []Chunk
}

const maxChunks = 4096

type FixedSizeChunker struct {
	Size    int
	Overlap int
}

func NewFixedSizeChunker(size, overlap int) *FixedSizeChunker {
	return &FixedSizeChunker{Size: size, Overlap: overlap}
}

func (c *FixedSizeChunker) Name() string { return "fixed" }

func (c *FixedSizeChunker) Chunk(text string) []Chunk {
	if text == "" {
		return []Chunk{{Index: 0, Text: ""}}
	}
	size := c.Size
	if size <= 0 {
		size = len(text)
	}
	overlap := c.Overlap
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= size {
		overlap = size / 8
	}
	if len(text) <= size {
		end := len(text)
		return []Chunk{{Index: 0, Text: text, Start: intPtr(0), End: &end}}
	}
	chunks := make([]Chunk, 0, min((len(text)+size-1)/max(size-overlap, 1), maxChunks))
	index := 0
	start := 0
	for start < len(text) && index < maxChunks {
		end := start + size
		if end > len(text) {
			end = len(text)
		}
		chunkText := text[start:end]
		s, e := start, end
		chunks = append(chunks, Chunk{Index: index, Text: chunkText, Start: &s, End: &e})
		index++
		if end >= len(text) {
			break
		}
		next := end - overlap
		if next <= start {
			next = end
		}
		start = next
	}
	return chunks
}

func intPtr(v int) *int { return &v }
