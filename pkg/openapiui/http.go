package openapiui

import (
	"fmt"
	"net/http"
	"path/filepath"
)

func Mount(mux *http.ServeMux, specDir string) {
	mux.HandleFunc("/openapi/account.json", serveFile(filepath.Join(specDir, "account", "auth.swagger.json")))
	mux.HandleFunc("/openapi/task.json", serveFile(filepath.Join(specDir, "task", "task.swagger.json")))
	mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/swagger", "/swagger/":
			http.Redirect(w, r, "/swagger/account", http.StatusFound)
		case "/swagger/account":
			serveSwaggerUI("Account API", "/openapi/account.json", false)(w, r)
		case "/swagger/task":
			serveSwaggerUI("Task API", "/openapi/task.json", true)(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

func serveFile(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}

func serveSwaggerUI(title, specURL string, withBearerAuth bool) http.HandlerFunc {
	specSetup := "return spec;"
	if withBearerAuth {
		specSetup = `spec.securityDefinitions = spec.securityDefinitions || {};
        spec.securityDefinitions.BearerAuth = {
          type: "apiKey",
          name: "Authorization",
          in: "header",
          description: "Use: Bearer <JWT>"
        };
        spec.security = [{ BearerAuth: [] }];
        return spec;`
	}

	page := fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>%s</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    body { margin: 0; background: #faf7f2; }
    .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    fetch(%q)
      .then((resp) => resp.json())
      .then((spec) => {
        %s
      })
      .then((spec) => {
        window.ui = SwaggerUIBundle({
          spec: spec,
          dom_id: '#swagger-ui',
          deepLinking: true,
          presets: [SwaggerUIBundle.presets.apis],
        });
      });
  </script>
</body>
</html>`, title, specURL, specSetup)

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	}
}
