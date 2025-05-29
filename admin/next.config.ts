import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  reactStrictMode: true, // Enables React Strict Mode for better development experience
  swcMinify: true, // Uses SWC for faster minification
  eslint: {
    ignoreDuringBuilds: true,
  },
  webpack(config) {
    config.module.rules.push({
      test: /\.svg$/,
      use: ['@svgr/webpack'], // Enables SVG imports as React components
    });
    return config;
  },
};

export default nextConfig;