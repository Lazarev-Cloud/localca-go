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
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
    
    // Log only in development
    if (process.env.NODE_ENV === 'development') {
      console.log(`API URL configuration: ${apiUrl}`);
    }
    
    return apiUrl;
  })(),
};

export default config; 