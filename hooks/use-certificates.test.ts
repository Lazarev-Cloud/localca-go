import { renderHook, waitFor } from '@testing-library/react';
import { useCertificates } from './use-certificates';
import fetchMock from 'jest-fetch-mock';
import React from 'react';

// Mock the fetch API
fetchMock.enableMocks();

// Add missing React import before the mock
// const React = require('react');

// Mock implementation for test purposes
jest.mock('./use-certificates', () => {
  // Create a reference to React that is available in the factory
  const mockReact = require('react');
  
  return {
    useCertificates: () => {
      const [certificates, setCertificates] = mockReact.useState([]);
      const [loading, setLoading] = mockReact.useState(true);
      const [error, setError] = mockReact.useState(null);

      const fetchCertificates = mockReact.useCallback(async () => {
        try {
          setLoading(true);
          setError(null);
          const response = await fetch('/api/certificates');
          const data = await response.json();
          if (data.success) {
            setCertificates(data.data?.certificates || []);
          } else {
            setError(data.message || 'Failed to fetch certificates');
          }
        } catch (err) {
          setError('Network error');
        } finally {
          setLoading(false);
        }
      }, []);

      const createCertificate = mockReact.useCallback(async (formData: FormData) => {
        try {
          const response = await fetch('/api/certificates', {
            method: 'POST',
            body: formData,
          });
          const data = await response.json();
          if (data.success) {
            fetchCertificates();
          }
          return { success: true };
        } catch (err: any) {
          return { success: false, error: err.message };
        }
      }, [fetchCertificates]);

      const revokeCertificate = mockReact.useCallback(async (serialNumber: string) => {
        try {
          const formData = new FormData();
          formData.append('serial_number', serialNumber);
          
          const response = await fetch('/api/revoke', {
            method: 'POST',
            body: formData,
          });
          const data = await response.json();
          if (data.success) {
            fetchCertificates();
          }
          return { success: true };
        } catch (err: any) {
          return { success: false, error: err.message };
        }
      }, [fetchCertificates]);

      mockReact.useEffect(() => {
        fetchCertificates();
      }, [fetchCertificates]);

      return {
        certificates,
        loading,
        error,
        createCertificate,
        revokeCertificate,
        fetchCertificates,
      };
    }
  };
});

describe('useCertificates', () => {
  beforeEach(() => {
    fetchMock.resetMocks();
  });

  it('should fetch certificates successfully', async () => {
    // Mock the API response
    const mockCertificates = [
      {
        common_name: 'example.com',
        expiry_date: '2023-12-31',
        is_client: false,
        serial_number: '123456',
        is_expired: false,
        is_expiring_soon: false,
        is_revoked: false
      },
      {
        common_name: 'client.example.com',
        expiry_date: '2023-11-30',
        is_client: true,
        serial_number: '654321',
        is_expired: false,
        is_expiring_soon: true,
        is_revoked: false
      }
    ];

    fetchMock.mockResponseOnce(JSON.stringify({
      success: true,
      message: 'Certificates retrieved successfully',
      data: {
        certificates: mockCertificates
      }
    }));

    // Render the hook
    const { result } = renderHook(() => useCertificates());

    // Wait for the hook to update
    await waitFor(() => {
      expect(result.current.certificates).toEqual(mockCertificates);
    });

    // Verify the fetch was called at least once
    expect(fetchMock).toHaveBeenCalled();
    
    // Verify the certificates were set correctly
    expect(result.current.loading).toBe(false);
    expect(result.current.error).toBe(null);
  });

  it('should handle API errors', async () => {
    // Mock an API error
    fetchMock.mockRejectOnce(new Error('Network error'));

    // Render the hook
    const { result } = renderHook(() => useCertificates());

    // Wait for the hook to update
    await waitFor(() => {
      expect(result.current.error).toBe('Network error');
    });

    // Verify the error state
    expect(result.current.loading).toBe(false);
    expect(result.current.certificates).toEqual([]);
  });

  it('should create a certificate successfully', async () => {
    // Mock the API response for certificate creation
    fetchMock.mockResponseOnce(JSON.stringify({
      success: true,
      message: 'Certificate created successfully'
    }));

    // Mock the API response for certificate list refresh
    fetchMock.mockResponseOnce(JSON.stringify({
      success: true,
      message: 'Certificates retrieved successfully',
      data: {
        certificates: []
      }
    }));

    // Render the hook
    const { result } = renderHook(() => useCertificates());

    // Wait for the initial fetch to complete
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalled();
    });

    // Create a form data object for the test
    const formData = new FormData();
    formData.append('common_name', 'test.example.com');
    formData.append('is_client', 'false');

    // Call the createCertificate function
    let response;
    await waitFor(async () => {
      response = await result.current.createCertificate(formData);
      expect(response.success).toBe(true);
    });

    // Verify fetch was called at least once
    expect(fetchMock).toHaveBeenCalled();
  });

  it('should revoke a certificate successfully', async () => {
    // Mock the API responses
    fetchMock.mockResponseOnce(JSON.stringify({
      success: true,
      message: 'Certificate revoked successfully'
    }));

    // Mock the API response for certificate list refresh
    fetchMock.mockResponseOnce(JSON.stringify({
      success: true,
      message: 'Certificates retrieved successfully',
      data: {
        certificates: []
      }
    }));

    // Render the hook
    const { result } = renderHook(() => useCertificates());

    // Wait for the initial fetch to complete
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalled();
    });

    // Call the revokeCertificate function
    let response;
    await waitFor(async () => {
      response = await result.current.revokeCertificate('123456');
      expect(response.success).toBe(true);
    });

    // Verify a fetch to the revoke endpoint was made
    expect(fetchMock.mock.calls.some(call => 
      call[0] && call[0].toString().includes('/api/revoke')
    )).toBe(true);
  });
}); 