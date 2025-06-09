import { renderHook, waitFor } from '@testing-library/react';
import { useRouter } from 'next/navigation';
import { useApi, ApiErrorType } from './use-api';
import fetchMock from 'jest-fetch-mock';

// Mock next/navigation
jest.mock('next/navigation', () => ({
  useRouter: jest.fn(),
}));

const mockPush = jest.fn();
const mockRouter = {
  push: mockPush,
  replace: jest.fn(),
  back: jest.fn(),
  forward: jest.fn(),
  refresh: jest.fn(),
  prefetch: jest.fn(),
};

describe('useApi', () => {
  beforeEach(() => {
    fetchMock.resetMocks();
    mockPush.mockClear();
    (useRouter as jest.Mock).mockReturnValue(mockRouter);
  });

  describe('fetchApi', () => {
    it('should make successful API request', async () => {
      const mockData = { success: true, message: 'Success', data: { test: 'value' } };
      fetchMock.mockResponseOnce(JSON.stringify(mockData));

      const { result } = renderHook(() => useApi());

      const response = await result.current.fetchApi('/test');

      expect(fetchMock).toHaveBeenCalledWith('/api/proxy/test', {
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        signal: expect.any(AbortSignal),
      });
      expect(response).toEqual(mockData);
      expect(result.current.loading).toBe(false);
      expect(result.current.error).toBe(null);
    });

    it('should handle server errors with retries', async () => {
      fetchMock
        .mockRejectOnce(new Error('Network error'))
        .mockRejectOnce(new Error('Network error'))
        .mockResponseOnce(JSON.stringify({ success: true, message: 'Success' }));

      const { result } = renderHook(() => useApi());

      const response = await result.current.fetchApi('/test');

      expect(fetchMock).toHaveBeenCalledTimes(3);
      expect(response.success).toBe(true);
    });

    it('should handle setup required response', async () => {
      const setupResponse = {
        success: false,
        message: 'Setup required',
        setupRequired: true
      };
      fetchMock.mockResponseOnce(JSON.stringify(setupResponse), { status: 401 });

      const { result } = renderHook(() => useApi());

      const response = await result.current.fetchApi('/test');

      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/setup');
      });
      expect(response.setupRequired).toBe(true);
      expect(result.current.error?.setupRequired).toBe(true);
    });

    it('should handle timeout errors', async () => {
      fetchMock.mockAbortOnce();

      const { result } = renderHook(() => useApi());

      const response = await result.current.fetchApi('/test');

      expect(response.success).toBe(false);
      expect(result.current.error?.type).toBe(ApiErrorType.TIMEOUT);
    });

    it('should handle network errors', async () => {
      fetchMock.mockRejectOnce(new TypeError('Network error'));

      const { result } = renderHook(() => useApi());

      const response = await result.current.fetchApi('/test');

      expect(response.success).toBe(false);
      expect(result.current.error?.type).toBe(ApiErrorType.NETWORK);
    });

    it('should handle HTTP error responses', async () => {
      const errorResponse = { message: 'Bad request' };
      fetchMock.mockResponseOnce(JSON.stringify(errorResponse), { status: 400 });

      const { result } = renderHook(() => useApi());

      const response = await result.current.fetchApi('/test');

      expect(response.success).toBe(false);
      expect(result.current.error?.status).toBe(400);
      expect(result.current.error?.type).toBe(ApiErrorType.SERVER);
    });
  });

  describe('postFormData', () => {
    it('should post form data successfully', async () => {
      const mockData = { success: true, message: 'Created' };
      fetchMock.mockResponseOnce(JSON.stringify(mockData));

      const { result } = renderHook(() => useApi());
      const formData = new FormData();
      formData.append('test', 'value');

      const response = await result.current.postFormData('/create', formData);

      expect(fetchMock).toHaveBeenCalledWith('/api/proxy/create', {
        method: 'POST',
        body: formData,
        signal: expect.any(AbortSignal),
      });
      expect(response).toEqual(mockData);
    });

    it('should handle form data upload errors', async () => {
      fetchMock.mockResponseOnce('Server Error', { status: 500 });

      const { result } = renderHook(() => useApi());
      const formData = new FormData();

      const response = await result.current.postFormData('/create', formData);

      expect(response.success).toBe(false);
      expect(result.current.error?.status).toBe(500);
    });

    it('should retry form data uploads on server errors', async () => {
      fetchMock
        .mockResponseOnce('Server Error', { status: 500 })
        .mockResponseOnce(JSON.stringify({ success: true, message: 'Success' }));

      const { result } = renderHook(() => useApi());
      const formData = new FormData();

      const response = await result.current.postFormData('/create', formData);

      expect(fetchMock).toHaveBeenCalledTimes(2);
      expect(response.success).toBe(true);
    });
  });

  describe('get method', () => {
    it('should make GET request', async () => {
      const mockData = { success: true, data: { items: [] } };
      fetchMock.mockResponseOnce(JSON.stringify(mockData));

      const { result } = renderHook(() => useApi());

      const response = await result.current.get('/items');

      expect(fetchMock).toHaveBeenCalledWith('/api/proxy/items', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        signal: expect.any(AbortSignal),
      });
      expect(response).toEqual(mockData);
    });
  });

  describe('post method', () => {
    it('should make POST request with JSON data', async () => {
      const mockData = { success: true, message: 'Created' };
      const postData = { name: 'test' };
      fetchMock.mockResponseOnce(JSON.stringify(mockData));

      const { result } = renderHook(() => useApi());

      const response = await result.current.post('/create', postData);

      expect(fetchMock).toHaveBeenCalledWith('/api/proxy/create', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        signal: expect.any(AbortSignal),
        body: JSON.stringify(postData),
      });
      expect(response).toEqual(mockData);
    });
  });

  describe('put method', () => {
    it('should make PUT request', async () => {
      const mockData = { success: true, message: 'Updated' };
      const putData = { id: 1, name: 'updated' };
      fetchMock.mockResponseOnce(JSON.stringify(mockData));

      const { result } = renderHook(() => useApi());

      const response = await result.current.put('/update/1', putData);

      expect(fetchMock).toHaveBeenCalledWith('/api/proxy/update/1', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        signal: expect.any(AbortSignal),
        body: JSON.stringify(putData),
      });
      expect(response).toEqual(mockData);
    });
  });

  describe('delete method', () => {
    it('should make DELETE request', async () => {
      const mockData = { success: true, message: 'Deleted' };
      fetchMock.mockResponseOnce(JSON.stringify(mockData));

      const { result } = renderHook(() => useApi());

      const response = await result.current.delete('/delete/1');

      expect(fetchMock).toHaveBeenCalledWith('/api/proxy/delete/1', {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        signal: expect.any(AbortSignal),
      });
      expect(response).toEqual(mockData);
    });
  });

  describe('loading state', () => {
    it('should set loading state during request', async () => {
      fetchMock.mockResponseOnce(JSON.stringify({ success: true }));

      const { result } = renderHook(() => useApi());

      // Start request
      const promise = result.current.fetchApi('/test');
      
      // Check loading state is true during request
      expect(result.current.loading).toBe(true);

      // Wait for completion
      await promise;

      // Check loading state is false after completion
      expect(result.current.loading).toBe(false);
    });
  });

  describe('error handling', () => {
    it('should clear previous errors on new request', async () => {
      // First request fails
      fetchMock.mockRejectOnce(new Error('First error'));
      
      const { result } = renderHook(() => useApi());
      
      await result.current.fetchApi('/test');
      expect(result.current.error).not.toBe(null);

      // Second request succeeds
      fetchMock.mockResponseOnce(JSON.stringify({ success: true }));
      
      await result.current.fetchApi('/test2');
      expect(result.current.error).toBe(null);
    });

    it('should handle malformed JSON responses', async () => {
      fetchMock.mockResponseOnce('Invalid JSON', { status: 500 });

      const { result } = renderHook(() => useApi());

      const response = await result.current.fetchApi('/test');

      expect(response.success).toBe(false);
      expect(result.current.error?.type).toBe(ApiErrorType.SERVER);
    });
  });
}); 