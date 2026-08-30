package spider_test

import (
	"testing"

	"github.com/spider/spider/pkg/security/chunking"
	"github.com/stretchr/testify/require"
)

func TestFixedSizeChunkerSingleChunk(t *testing.T) {
	c := chunking.NewFixedSizeChunker(2048, 128)
	chunks := c.Chunk("hello")
	require.Len(t, chunks, 1)
	require.Equal(t, "hello", chunks[0].Text)
}

func TestFixedSizeChunkerSplitsLongText(t *testing.T) {
	c := chunking.NewFixedSizeChunker(10, 2)
	text := "01234567890123456789"
	chunks := c.Chunk(text)
	require.Greater(t, len(chunks), 1)
}

func TestFixedSizeChunkerZeroSizeDoesNotLoop(t *testing.T) {
	c := chunking.NewFixedSizeChunker(0, 0)
	chunks := c.Chunk("Explain distributed systems")
	require.Len(t, chunks, 1)
	require.Equal(t, "Explain distributed systems", chunks[0].Text)
}
