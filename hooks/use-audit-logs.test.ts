import { renderHook, waitFor } from '@testing-library/react';
import { useAuditLogs } from './use-audit-logs';

// Mock the useApi hook
const mockFetchApi = jest.fn();
jest.mock('./use-api', () => ({
  useApi: () => ({
    fetchApi: mockFetchApi,
    loading: false,
    error: null,
  }),
}));

// Reset global cache before each test
const resetGlobalCache = () => {
  // Access the global cache through the module
  const auditLogsModule = require('./use-audit-logs');
  if (auditLogsModule.globalCache) {
    auditLogsModule.globalCache.data = null;
    auditLogsModule.globalCache.timestamp = 0;
    auditLogsModule.globalCache.loading = false;
  }
};

describe('useAuditLogs', () => {
  beforeEach(() => {
    mockFetchApi.mockClear();
    resetGlobalCache();
    jest.clearAllMocks();
  });

  it('should fetch audit logs successfully', async () => {
    const mockLogs = [
      {
        id: 1,
        action: 'certificate_created',
        resource: 'certificate',
        resource_id: 'cert-123',
        user_ip: '192.168.1.1',
        user_agent: 'Mozilla/5.0',
        details: 'Created certificate for example.com',
        success: true,
        created_at: '2024-01-01T10:00:00Z',
      },
      {
        id: 2,
        action: 'certificate_revoked',
        resource: 'certificate',
        resource_id: 'cert-456',
        user_ip: '192.168.1.1',
        user_agent: 'Mozilla/5.0',
        details: 'Revoked certificate for test.com',
        success: true,
        created_at: '2024-01-01T11:00:00Z',
      },
    ];

    mockFetchApi.mockResolvedValueOnce({
      success: true,
      data: {
        audit_logs: mockLogs,
        total: 2,
        limit: 10,
        offset: 0,
      },
    });

    const { result } = renderHook(() => useAuditLogs());

    await waitFor(() => {
      expect(result.current.auditLogs).toEqual(mockLogs);
    });

    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBe(null);
    expect(mockFetchApi).toHaveBeenCalledWith('/api/audit-logs?limit=50&offset=0');
  });

  it('should handle API errors', async () => {
    mockFetchApi.mockResolvedValueOnce({
      success: false,
      message: 'Network error'
    });

    const { result } = renderHook(() => useAuditLogs());

    await waitFor(() => {
      expect(result.current.error).toBe('Failed to load audit logs');
    });

    expect(result.current.loading).toBe(false);
    expect(result.current.auditLogs).toEqual([]);
  });

  it('should handle pagination with limit and offset', async () => {
    const mockLogs = Array.from({ length: 5 }, (_, i) => ({
      id: i + 1,
      action: 'certificate_created',
      resource: 'certificate',
      resource_id: `cert-${i + 1}`,
      user_ip: '192.168.1.1',
      user_agent: 'Mozilla/5.0',
      details: `Created certificate ${i + 1}`,
      success: true,
      created_at: `2024-01-01T${10 + i}:00:00Z`,
    }));

    mockFetchApi.mockResolvedValueOnce({
      success: true,
      data: {
        audit_logs: mockLogs,
        total: 50,
        limit: 5,
        offset: 0,
      },
    });

    const { result } = renderHook(() => useAuditLogs(5, 0));

    await waitFor(() => {
      expect(result.current.auditLogs).toHaveLength(5);
    });

    expect(result.current.loading).toBe(false);
    expect(mockFetchApi).toHaveBeenCalledWith('/api/audit-logs?limit=50&offset=0');
  });

  it('should refresh logs when called', async () => {
    const mockLogs = [
      {
        id: 1,
        action: 'certificate_created',
        resource: 'certificate',
        resource_id: 'cert-123',
        user_ip: '192.168.1.1',
        user_agent: 'Mozilla/5.0',
        details: 'Created certificate for example.com',
        success: true,
        created_at: '2024-01-01T10:00:00Z',
      },
    ];

    const updatedLogs = [
      ...mockLogs,
      {
        id: 2,
        action: 'certificate_revoked',
        resource: 'certificate',
        resource_id: 'cert-456',
        user_ip: '192.168.1.1',
        user_agent: 'Mozilla/5.0',
        details: 'Revoked certificate for test.com',
        success: true,
        created_at: '2024-01-01T11:00:00Z',
      },
    ];

    mockFetchApi
      .mockResolvedValueOnce({
        success: true,
        data: { audit_logs: mockLogs, total: 1, limit: 10, offset: 0 },
      })
      .mockResolvedValueOnce({
        success: true,
        data: { audit_logs: updatedLogs, total: 2, limit: 10, offset: 0 },
      });

    const { result } = renderHook(() => useAuditLogs());

    await waitFor(() => {
      expect(result.current.auditLogs).toHaveLength(1);
    });

    // Refresh logs
    result.current.refresh();

    await waitFor(() => {
      expect(result.current.auditLogs).toHaveLength(2);
    });

    expect(mockFetchApi).toHaveBeenCalledTimes(2);
  });

  it('should handle empty response', async () => {
    mockFetchApi.mockResolvedValueOnce({
      success: true,
      data: {
        audit_logs: [],
        total: 0,
        limit: 10,
        offset: 0,
      },
    });

    const { result } = renderHook(() => useAuditLogs());

    await waitFor(() => {
      expect(result.current.auditLogs).toEqual([]);
    });

    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBe(null);
  });

  it('should handle malformed response', async () => {
    mockFetchApi.mockResolvedValueOnce({
      success: false,
      message: 'Invalid request',
    });

    const { result } = renderHook(() => useAuditLogs());

    await waitFor(() => {
      expect(result.current.error).toBe('Failed to load audit logs');
    });

    expect(result.current.auditLogs).toEqual([]);
    expect(result.current.loading).toBe(false);
  });

  it('should handle different limit and offset parameters', async () => {
    const mockLogs = Array.from({ length: 3 }, (_, i) => ({
      id: i + 6,
      action: 'certificate_created',
      resource: 'certificate',
      resource_id: `cert-${i + 6}`,
      user_ip: '192.168.1.1',
      user_agent: 'Mozilla/5.0',
      details: `Created certificate ${i + 6}`,
      success: true,
      created_at: `2024-01-01T${15 + i}:00:00Z`,
    }));

    mockFetchApi.mockResolvedValueOnce({
      success: true,
      data: {
        audit_logs: mockLogs,
        total: 3,
        limit: 50,
        offset: 0,
      },
    });

    const { result } = renderHook(() => useAuditLogs(50, 0));

    await waitFor(() => {
      expect(result.current.auditLogs).toHaveLength(3);
    });

    expect(mockFetchApi).toHaveBeenCalledWith('/api/audit-logs?limit=50&offset=0');
  });
}); 