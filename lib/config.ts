// Configuration for the application
interface Config {
  apiUrl: string;
}

const config: Config = {
  // API URL for backend
  apiUrl: (() => {
    // For client-side, use relative URLs to proxy through Next.js
    if (typeof window !== 'undefined') {
      return '';
    }
    
    // For server-side rendering, use the internal Docker network URL
    // In Docker, backend service is accessible at 'backend:8080'
    // In development, use localhost:8080
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
    
    // Log the API URL configuration for debugging
    console.log(`[Config] API URL: ${apiUrl}, NODE_ENV: ${process.env.NODE_ENV}`);
    
    return apiUrl;
  })(),
};

export default config; 