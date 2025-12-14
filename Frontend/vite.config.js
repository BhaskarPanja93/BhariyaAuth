import {defineConfig} from 'vite'
import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import compression from 'vite-plugin-compression2'

// https://vite.dev/config/
export default defineConfig({
    base: '/auth/',
    plugins: [
        react(),
        tailwindcss(),
        compression({
            algorithm: 'gzip', ext: '.gz', threshold: 1024, deleteOriginFile: false
        }),
        compression({
            algorithm: 'brotliCompress', ext: '.br', threshold: 1024, deleteOriginFile: false
        })],
})
