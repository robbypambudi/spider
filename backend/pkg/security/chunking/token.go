package chunking

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SidecarTokenChunker splits text via the prompt-shield sidecar /chunk endpoint
// using the same tokenizer as lab eval_chunked.py (unit=tokens).
type SidecarTokenChunker struct {
	Endpoint     string
	ChunkSize    int
	Overlap      int
	HTTPClient   *http.Client
}

type chunkRequest struct {
	Text       string `json:"text"`
	ChunkSize  int    `json:"chunk_size"`
	Overlap    int    `json:"overlap"`
	Unit       string `json:"unit"`
}

type chunkItemResponse struct {
	Index int    `json:"index"`
	Text  string `json:"text"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

type chunkResponse struct {
	Chunks []chunkItemResponse `json:"chunks"`
}

func NewSidecarTokenChunker(endpoint string, chunkSize, overlap int) *SidecarTokenChunker {
	if chunkSize <= 0 {
		chunkSize = 256
	}
	if overlap < 0 {
		overlap = 0
	}
	return &SidecarTokenChunker{
		Endpoint:  endpoint,
		ChunkSize: chunkSize,
		Overlap:   overlap,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *SidecarTokenChunker) Name() string { return "token" }

func (c *SidecarTokenChunker) Chunk(text string) []Chunk {
	if text == "" {
		return []Chunk{{Index: 0, Text: ""}}
	}
	body, _ := json.Marshal(chunkRequest{
		Text:      text,
		ChunkSize: c.ChunkSize,
		Overlap:   c.Overlap,
		Unit:      "tokens",
	})
	resp, err := c.HTTPClient.Post(c.Endpoint+"/chunk", "application/json", bytes.NewReader(body))
	if err != nil {
		return fallbackFixedChunk(text, c.ChunkSize, c.Overlap)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fallbackFixedChunk(text, c.ChunkSize, c.Overlap)
	}
	var result chunkResponse
	if err := json.Unmarshal(raw, &result); err != nil || len(result.Chunks) == 0 {
		return fallbackFixedChunk(text, c.ChunkSize, c.Overlap)
	}
	if len(result.Chunks) > maxChunks {
		result.Chunks = result.Chunks[:maxChunks]
	}
	out := make([]Chunk, 0, len(result.Chunks))
	for _, item := range result.Chunks {
		start, end := item.Start, item.End
		out = append(out, Chunk{
			Index: item.Index,
			Text:  item.Text,
			Start: &start,
			End:   &end,
			Metadata: map[string]interface{}{
				"unit": "tokens",
			},
		})
	}
	return out
}

func fallbackFixedChunk(text string, size, overlap int) []Chunk {
	return NewFixedSizeChunker(size*4, overlap*4).Chunk(text)
}

// ChunkSequence mirrors lab create_chunk_datasets.chunk_sequence for tests.
func ChunkSequence(n, size, overlap int) [][2]int {
	if n == 0 {
		return [][2]int{{0, 0}}
	}
	if size <= 0 || n <= size {
		return [][2]int{{0, n}}
	}
	step := size - overlap
	if step < 1 {
		step = 1
	}
	var spans [][2]int
	start := 0
	for start < n && len(spans) < maxChunks {
		end := start + size
		if end > n {
			end = n
		}
		spans = append(spans, [2]int{start, end})
		if end >= n {
			break
		}
		start += step
	}
	return spans
}

func (c *SidecarTokenChunker) String() string {
	return fmt.Sprintf("token(size=%d,overlap=%d)", c.ChunkSize, c.Overlap)
}
