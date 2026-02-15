import { defineConfig } from "vite";

export default defineConfig({
  server: {
    port: 5173,
    proxy: {
      "/configs": "http://localhost:8080",
      "/health": "http://localhost:8080",
      "/doc.json": "http://localhost:8080",
      "/doc.yaml": "http://localhost:8080"
    }
  },
  test: {
    environment: "node"
  }
});

