window.onload = function() {
  window.ui = SwaggerUIBundle({
    url: "/api/docs/openapi.yml", // Ensure this matches the route in your Go server
    dom_id: '#swagger-ui',
    presets: [
      SwaggerUIBundle.presets.apis,
      SwaggerUIStandalonePreset
    ],
    layout: "BaseLayout"
  });
};