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
    // Only add rewrites if we're not using the proxy routes
    // The proxy routes handle API forwarding internally
    if (process.env.USE_PROXY_ROUTES === 'false') {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
      
      return [
        {
          source: '/api/:path*',
          destination: `${apiUrl}/api/:path*`,
        },
      ];
    }
    
    return [];
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