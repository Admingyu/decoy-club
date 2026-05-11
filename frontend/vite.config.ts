import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";

const backendTarget = "http://192.168.191.178:8080";

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 3000,
    host: "0.0.0.0",
    proxy: {
      "/api": {
        target: backendTarget,
        changeOrigin: true,
      },
      "/uploads": {
        target: backendTarget,
        changeOrigin: true,
      },
    },
  },
});
