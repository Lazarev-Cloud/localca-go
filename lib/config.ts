// Configuration for the application
interface Config {
  apiUrl: string;
}

const config: Config = {
  // API URL for backend
  apiUrl: (() => {
    // For client-side, use empty string to make relative requests
    if (typeof window !== 'undefined') {
      return '';
    }
    
    // For server-side, check environment variable or use default
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
    
    // Log only in development
    if (process.env.NODE_ENV === 'development') {
      console.log(`API URL configuration: ${apiUrl}`);
    }
    
    return apiUrl;
  })(),
};

export default config; 