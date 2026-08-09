import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [react()],
  build: { sourcemap: false },
  server: {
    proxy: {
      "/widgets": "http://localhost:3000",
    },
  },
});
