declare global {
  var testConfig: {
    backendUrl: string;
    timeout: number;
    retryAttempts: number;
    retryDelay: number;
  };

  var waitForBackend: (url?: string, timeout?: number) => Promise<boolean>;
  var resetBackendState: () => Promise<void>;
  var __BACKEND_PID__: number | undefined;
}

export {}; 