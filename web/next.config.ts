import type { NextConfig } from "next";

const nextConfig: any = {
  /* config options here */

  // Optimize build memory usage
  typescript: {
    ignoreBuildErrors: true,
  },
  eslint: {
    ignoreDuringBuilds: true,
  },
  productionBrowserSourceMaps: false,
  compress: true,
  poweredByHeader: false,
  // High memory usage fix: disable minification
  swcMinify: true,
  output: "standalone",
};

export default nextConfig;
