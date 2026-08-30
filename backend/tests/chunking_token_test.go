package spider_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spider/spider/pkg/security/chunking"
	"github.com/stretchr/testify/require"
)

func TestChunkSequenceSpans(t *testing.T) {
	spans := chunking.ChunkSequence(1000, 256, 0)
	require.Greater(t, len(spans), 1)
	require.Equal(t, [2]int{0, 256}, spans[0])
	require.Equal(t, 256, spans[1][0])
}

func TestSidecarTokenChunkerUsesEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/chunk", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"chunks": []map[string]interface{}{
				{"index": 0, "text": "hello", "start": 0, "end": 5},
			},
		})
	}))
	defer server.Close()

	c := chunking.NewSidecarTokenChunker(server.URL, 256, 0)
	chunks := c.Chunk("hello world")
	require.Len(t, chunks, 1)
	require.Equal(t, "hello", chunks[0].Text)
	require.Equal(t, "token", c.Name())
}

func TestSidecarTokenChunkerFallbackOnError(t *testing.T) {
	c := chunking.NewSidecarTokenChunker("http://127.0.0.1:1", 256, 0)
	chunks := c.Chunk("short text")
	require.Len(t, chunks, 1)
	require.Equal(t, "short text", chunks[0].Text)
}
