package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DocHandler serves the OpenAPI documentation using Redocly
func DocHandler(c *gin.Context) {
	// Read the openapi.yaml file from the project root directory
	data, err := ioutil.ReadFile("openapi.yaml")
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to read openapi.yaml: %v", err))
		return
	}

	// Embed the YAML content into an HTML page using Redocly CDN
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
  <head>
    <title>API Documentation</title>
    <!-- Needed for adaptive design -->
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <!-- Redocly CDN -->
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
  </head>
  <body>
    <redoc spec='%s'></redoc>
  </body>
</html>`, string(data))

	// Serve the HTML page
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}