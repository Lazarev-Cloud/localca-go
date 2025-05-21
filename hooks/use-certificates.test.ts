import { renderHook, waitFor } from '@testing-library/react';
import { useCertificates } from './use-certificates';
import fetchMock from 'jest-fetch-mock';

// Mock the fetch API
fetchMock.enableMocks();

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

    // Verify the fetch was called correctly
    expect(fetchMock).toHaveBeenCalledWith('/api/certificates', expect.any(Object));
    
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
      expect(fetchMock).toHaveBeenCalledWith('/api/certificates', expect.any(Object));
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

    // Verify the fetch calls
    expect(fetchMock).toHaveBeenCalledWith('/api/certificates', {
      method: 'POST',
      body: formData,
    });
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
      expect(fetchMock).toHaveBeenCalledWith('/api/certificates', expect.any(Object));
    });

    // Call the revokeCertificate function
    let response;
    await waitFor(async () => {
      response = await result.current.revokeCertificate('123456');
      expect(response.success).toBe(true);
    });

    // Verify the fetch calls
    expect(fetchMock).toHaveBeenCalledWith('/api/revoke', {
      method: 'POST',
      body: expect.any(FormData),
    });
  });
}); 