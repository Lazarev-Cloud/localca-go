import { codecovNextJSWebpackPlugin } from "@codecov/nextjs-webpack-plugin";

/** @type {import('next').NextConfig} */
const nextConfig = {
  eslint: {
    ignoreDuringBuilds: true,
  },
  typescript: {
    ignoreBuildErrors: true,
  },
  images: {
    unoptimized: true,
  },
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://backend:8080/api/:path*',
      },
    ];
  },
  webpack: (config, options) => {
    // Add Codecov bundle analysis plugin
    if (process.env.NODE_ENV === 'production' && process.env.CODECOV_TOKEN) {
      config.plugins.push(
        codecovNextJSWebpackPlugin({
          enableBundleAnalysis: true,
          bundleName: "localca-go-frontend-bundle",
          uploadToken: process.env.CODECOV_TOKEN,
          webpack: options.webpack,
        }),
      );
    }
    
    // Return the modified config
    return config;
  },
}

export default nextConfig