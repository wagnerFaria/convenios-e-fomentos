package httpapi

import "net/http"

// A UI da documentação é carregada via CDN, apenas para leitura.
const docsPage = `<!doctype html>
<html lang="pt-BR">
<head>
  <meta charset="utf-8">
  <title>Documentação da API</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
<body>
  <script id="api-reference" data-url="/api/docs/openapi.yaml"></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.25.0"></script>
</body>
</html>`

type docs struct{ openapi []byte }

func docsHandler(spec []byte) docs { return docs{openapi: spec} }

func (d docs) page(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// CSP mais permissiva só para esta página (CDN do Scalar).
	w.Header().Set("Content-Security-Policy",
		"default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://fonts.googleapis.com; font-src https://fonts.gstatic.com https://cdn.jsdelivr.net; img-src 'self' data:; connect-src 'self'")
	_, _ = w.Write([]byte(docsPage))
}

func (d docs) spec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	_, _ = w.Write(d.openapi)
}
