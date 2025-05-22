// Configuration for the application
const config = {
  // API URL for backend
  apiUrl: (() => {
    // Get environment variable
    const configuredUrl = process.env.NEXT_PUBLIC_API_URL;
    
    // If it's specifically set, use it
    if (configuredUrl) {
      console.log(`Using configured API URL: ${configuredUrl}`);
      return configuredUrl;
    }
    
    // In a browser environment, use relative URL
    if (typeof window !== 'undefined') {
      // Use relative URL if in browser
      console.log('Using relative API URL for browser environment');
      return '';
    }
    
    // In server environment (but not configured), default to localhost
    console.log('Using default API URL: http://localhost:8080');
    return 'http://localhost:8080';
  })(),
};

console.log(`Final API URL configuration: ${config.apiUrl}`);
export default config; 