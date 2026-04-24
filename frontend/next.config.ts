import type { NextConfig } from "next";

const API_ORIGIN = process.env.API_ORIGIN ?? "http://localhost:8080";
const MINIO_PUBLIC = process.env.MINIO_PUBLIC_ORIGIN ?? "http://localhost:9000";
const isDev = process.env.NODE_ENV !== "production";

const nextConfig: NextConfig = {
  output: "standalone",
  async rewrites() {
    return [
      {
        source: "/api/:path*",
        destination: `${API_ORIGIN}/api/:path*`,
      },
    ];
  },

  async headers() {
    return [
      {
        source: "/(.*)",
        headers: [
          { key: "X-Frame-Options",           value: "DENY" },
          { key: "X-Content-Type-Options",     value: "nosniff" },
          { key: "Referrer-Policy",            value: "strict-origin-when-cross-origin" },
          { key: "Permissions-Policy",         value: "camera=(), microphone=(), geolocation=()" },
          {
            key: "Content-Security-Policy",
            // script-src 'self' 'unsafe-inline' required for Next.js inline scripts in App Router
            // 'unsafe-eval' is required in dev mode by React 19 for callstack reconstruction
            value: [
              "default-src 'self'",
              `script-src 'self' 'unsafe-inline'${isDev ? " 'unsafe-eval'" : ""}`,
              "style-src 'self' 'unsafe-inline'",
              `img-src 'self' data: https: ${MINIO_PUBLIC}`,
              `connect-src 'self' blob: ${MINIO_PUBLIC}`,
              "font-src 'self'",
              `frame-src 'self' ${MINIO_PUBLIC}`,
              "frame-ancestors 'none'",
              "worker-src 'self' blob:",
            ].join("; "),
          },
        ],
      },
    ];
  },
};

export default nextConfig;
