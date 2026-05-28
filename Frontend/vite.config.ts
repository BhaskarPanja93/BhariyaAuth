import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import compression, {defineAlgorithm} from 'vite-plugin-compression2'
import {constants} from "zlib";

export default defineConfig({
  base: '/auth/',
  plugins: [
    react(),
    tailwindcss(),
    compression({
      deleteOriginalAssets: false,
      algorithms: [
        defineAlgorithm('gzip', { level: 9 }),
        defineAlgorithm('brotliCompress', {
          params: {
            [constants.BROTLI_PARAM_QUALITY]: 11
          }
        })
      ]
    })
  ]
})

