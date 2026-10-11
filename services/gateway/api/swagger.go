package api

import _ "embed"

// Swagger UI 5.17.14 (Apache-2.0), vendored so /docs does not load scripts from unpkg.

//go:embed swagger/swagger-ui.css
var SwaggerCSS []byte

//go:embed swagger/swagger-ui-bundle.js
var SwaggerJS []byte

//go:embed swagger/init.js
var SwaggerInit []byte

// DocsHTML is the Swagger page. Assets are same-origin and sit behind the docs gate.
const DocsHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8"/>
  <title>Event Horizon API</title>
  <meta http-equiv="Content-Security-Policy" content="default-src 'none'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'"/>
  <link rel="stylesheet" href="/docs/swagger-ui.css"/>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="/docs/swagger-ui-bundle.js"></script>
  <script src="/docs/swagger-init.js"></script>
</body>
</html>`
