// Package docs serves the OpenAPI specification and an embedded Swagger UI.
package docs

import (
	"net/http"

	_ "embed"

	"github.com/go-chi/chi/v5"
)

//go:embed api-spec.yml
var spec []byte

// ponytail: Swagger UI assets come from the jsdelivr CDN instead of vendoring
// ~3MB of swagger-ui-dist into the repo. Vendor them if /docs must work offline.
const swaggerUI = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Spider API</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.17.14/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5.17.14/swagger-ui-bundle.js" crossorigin></script>
  <script>
    window.onload = () => SwaggerUIBundle({
      url: "/api-spec.yml",
      dom_id: "#swagger-ui",
      persistAuthorization: true,
    });
  </script>
</body>
</html>`

// Mount registers the Swagger UI at /docs and the raw spec at /api-spec.yml.
func Mount(r chi.Router) {
	r.Get("/docs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(swaggerUI))
	})
	r.Get("/api-spec.yml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(spec)
	})
}
