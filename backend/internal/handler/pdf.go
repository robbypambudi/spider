package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/spider/spider/internal/app"
	"github.com/spider/spider/internal/middleware"
	"github.com/spider/spider/internal/spidererrors"
)

const (
	maxPDFBytes = 10 << 20 // 10 MiB
	maxPDFChars = 1 << 20  // 1 MiB of extracted text is plenty for a scan
)

// extractPDFText pulls the plain text out of every page of an uncorrupted,
// unencrypted PDF. Pages that fail to decode are skipped rather than fatal.
func extractPDFText(data []byte) (text string, pages int, err error) {
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		return "", 0, spidererrors.Validation("File is not a PDF")
	}
	// ponytail: the parser panics on some malformed files instead of erroring.
	defer func() {
		if rec := recover(); rec != nil {
			text, pages, err = "", 0, spidererrors.Validation(fmt.Sprintf("Malformed PDF: %v", rec))
		}
	}()

	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", 0, spidererrors.Validation("Cannot read PDF (corrupt or password protected)")
	}

	var buf strings.Builder
	pages = reader.NumPage()
	for i := 1; i <= pages; i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		content, perr := page.GetPlainText(nil)
		if perr != nil {
			continue
		}
		buf.WriteString(content)
		buf.WriteString("\n")
		if buf.Len() >= maxPDFChars {
			break
		}
	}

	out := strings.TrimSpace(buf.String())
	if len(out) > maxPDFChars {
		out = out[:maxPDFChars]
	}
	if out == "" {
		return "", pages, spidererrors.Validation("No extractable text in PDF (scanned image?)")
	}
	return out, pages, nil
}

func scanPDFHandler(c *app.Container) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxPDFBytes)

		file, header, err := r.FormFile("file")
		if err != nil {
			WriteError(w, spidererrors.Validation("Expected a multipart/form-data upload with a 'file' field (max 10MB)"))
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			WriteError(w, spidererrors.Validation("Upload too large or truncated (max 10MB)"))
			return
		}

		text, pages, err := extractPDFText(data)
		if err != nil {
			WriteError(w, err)
			return
		}

		user, _ := middleware.UserFromContext(r.Context())
		meta := map[string]interface{}{
			"filename":  header.Filename,
			"pages":     pages,
			"size_byte": len(data),
		}
		var model *string
		if v := r.FormValue("model"); v != "" {
			model = &v
		}
		result, stored, err := c.Security.Inspect(r.Context(), text, inspectOpts(user, "pdf", model, meta, true))
		if err != nil {
			WriteError(w, err)
			return
		}
		var scanID *string
		if stored != nil {
			s := stored.ID.String()
			scanID = &s
		}
		WriteJSON(w, http.StatusOK, toScanResponse(result, scanID))
	}
}
