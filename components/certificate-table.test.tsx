import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CertificateTable } from './certificate-table';

// Mock the hooks
const mockCertificates = [
  {
    common_name: 'example.com',
    serial_number: '123456789',
    expiry_date: '2025-01-01T00:00:00Z',
    is_client: false,
    is_expired: false,
    is_expiring_soon: false,
    is_revoked: false,
  },
  {
    common_name: 'client.example.com',
    serial_number: '987654321',
    expiry_date: '2024-12-01T00:00:00Z',
    is_client: true,
    is_expired: false,
    is_expiring_soon: true,
    is_revoked: false,
  },
  {
    common_name: 'expired.example.com',
    serial_number: '555666777',
    expiry_date: '2023-01-01T00:00:00Z',
    is_client: false,
    is_expired: true,
    is_expiring_soon: false,
    is_revoked: false,
  },
];

const mockRevokeCertificate = jest.fn();
const mockRenewCertificate = jest.fn();
const mockDeleteCertificate = jest.fn();

jest.mock('@/hooks/use-certificates', () => ({
  useCertificates: () => ({
    certificates: mockCertificates,
    loading: false,
    error: null,
    revokeCertificate: mockRevokeCertificate,
    renewCertificate: mockRenewCertificate,
    deleteCertificate: mockDeleteCertificate,
  }),
}));

jest.mock('./certificate-filters', () => ({
  useCertificateFilters: () => ({
    filters: {
      searchQuery: '',
      certificateType: 'all',
      status: 'all',
    },
  }),
}));

jest.mock('@/hooks/use-toast-new', () => ({
  useToast: () => ({
    toast: jest.fn(),
  }),
}));

// Mock fetch for download functionality
global.fetch = jest.fn();
global.URL.createObjectURL = jest.fn(() => 'mock-url');
global.URL.revokeObjectURL = jest.fn();

// Mock DOM methods
Object.defineProperty(document, 'createElement', {
  value: jest.fn(() => ({
    href: '',
    download: '',
    click: jest.fn(),
  })),
});

Object.defineProperty(document.body, 'appendChild', {
  value: jest.fn(),
});

Object.defineProperty(document.body, 'removeChild', {
  value: jest.fn(),
});

describe('CertificateTable', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    (global.fetch as jest.Mock).mockClear();
  });

  it('should render certificate table with data', () => {
    render(<CertificateTable />);

    expect(screen.getByText('example.com')).toBeInTheDocument();
    expect(screen.getByText('client.example.com')).toBeInTheDocument();
    expect(screen.getByText('expired.example.com')).toBeInTheDocument();
  });

  it('should display correct certificate types', () => {
    render(<CertificateTable />);

    // Check for server certificate badge
    expect(screen.getByText('Server')).toBeInTheDocument();
    
    // Check for client certificate badge
    expect(screen.getByText('Client')).toBeInTheDocument();
  });

  it('should display correct certificate statuses', () => {
    render(<CertificateTable />);

    // Check for valid status
    expect(screen.getByText('Valid')).toBeInTheDocument();
    
    // Check for expiring soon status
    expect(screen.getByText('Expiring Soon')).toBeInTheDocument();
    
    // Check for expired status
    expect(screen.getByText('Expired')).toBeInTheDocument();
  });

  it('should handle certificate download', async () => {
    const mockBlob = new Blob(['certificate content'], { type: 'application/x-pem-file' });
    (global.fetch as jest.Mock).mockResolvedValueOnce({
      ok: true,
      blob: () => Promise.resolve(mockBlob),
    });

    render(<CertificateTable />);

    // Find and click the first dropdown menu
    const dropdownTriggers = screen.getAllByRole('button', { name: /open menu/i });
    fireEvent.click(dropdownTriggers[0]);

    // Click download option
    const downloadButton = screen.getByText('Download');
    fireEvent.click(downloadButton);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalledWith(
        '/api/proxy/api/download/example.com/crt',
        { credentials: 'include' }
      );
    });
  });

  it('should handle certificate renewal', async () => {
    mockRenewCertificate.mockResolvedValueOnce({ success: true });

    render(<CertificateTable />);

    // Find and click the first dropdown menu
    const dropdownTriggers = screen.getAllByRole('button', { name: /open menu/i });
    fireEvent.click(dropdownTriggers[0]);

    // Click renew option
    const renewButton = screen.getByText('Renew');
    fireEvent.click(renewButton);

    await waitFor(() => {
      expect(mockRenewCertificate).toHaveBeenCalledWith('123456789');
    });
  });

  it('should handle certificate revocation with confirmation', async () => {
    // Mock window.confirm
    const originalConfirm = window.confirm;
    window.confirm = jest.fn(() => true);

    mockRevokeCertificate.mockResolvedValueOnce({ success: true });

    render(<CertificateTable />);

    // Find and click the first dropdown menu
    const dropdownTriggers = screen.getAllByRole('button', { name: /open menu/i });
    fireEvent.click(dropdownTriggers[0]);

    // Click revoke option
    const revokeButton = screen.getByText('Revoke');
    fireEvent.click(revokeButton);

    await waitFor(() => {
      expect(window.confirm).toHaveBeenCalledWith(
        'Are you sure you want to revoke the certificate "example.com"? This action cannot be undone.'
      );
      expect(mockRevokeCertificate).toHaveBeenCalledWith('123456789');
    });

    // Restore original confirm
    window.confirm = originalConfirm;
  });

  it('should handle certificate deletion with confirmation', async () => {
    // Mock window.confirm
    const originalConfirm = window.confirm;
    window.confirm = jest.fn(() => true);

    mockDeleteCertificate.mockResolvedValueOnce({ success: true });

    render(<CertificateTable />);

    // Find and click the first dropdown menu
    const dropdownTriggers = screen.getAllByRole('button', { name: /open menu/i });
    fireEvent.click(dropdownTriggers[0]);

    // Click delete option
    const deleteButton = screen.getByText('Delete');
    fireEvent.click(deleteButton);

    await waitFor(() => {
      expect(window.confirm).toHaveBeenCalledWith(
        'Are you sure you want to delete the certificate "example.com"? This action cannot be undone.'
      );
      expect(mockDeleteCertificate).toHaveBeenCalledWith('123456789');
    });

    // Restore original confirm
    window.confirm = originalConfirm;
  });

  it('should not perform action when confirmation is cancelled', async () => {
    // Mock window.confirm to return false
    const originalConfirm = window.confirm;
    window.confirm = jest.fn(() => false);

    render(<CertificateTable />);

    // Find and click the first dropdown menu
    const dropdownTriggers = screen.getAllByRole('button', { name: /open menu/i });
    fireEvent.click(dropdownTriggers[0]);

    // Click revoke option
    const revokeButton = screen.getByText('Revoke');
    fireEvent.click(revokeButton);

    await waitFor(() => {
      expect(window.confirm).toHaveBeenCalled();
      expect(mockRevokeCertificate).not.toHaveBeenCalled();
    });

    // Restore original confirm
    window.confirm = originalConfirm;
  });

  it('should handle download errors gracefully', async () => {
    (global.fetch as jest.Mock).mockRejectedValueOnce(new Error('Download failed'));

    render(<CertificateTable />);

    // Find and click the first dropdown menu
    const dropdownTriggers = screen.getAllByRole('button', { name: /open menu/i });
    fireEvent.click(dropdownTriggers[0]);

    // Click download option
    const downloadButton = screen.getByText('Download');
    fireEvent.click(downloadButton);

    await waitFor(() => {
      expect(global.fetch).toHaveBeenCalled();
    });
  });

  it('should show loading state during actions', async () => {
    // Mock a slow response
    mockRenewCertificate.mockImplementationOnce(
      () => new Promise(resolve => setTimeout(() => resolve({ success: true }), 100))
    );

    render(<CertificateTable />);

    // Find and click the first dropdown menu
    const dropdownTriggers = screen.getAllByRole('button', { name: /open menu/i });
    fireEvent.click(dropdownTriggers[0]);

    // Click renew option
    const renewButton = screen.getByText('Renew');
    fireEvent.click(renewButton);

    // Check for loading state (spinner icon)
    expect(screen.getByTestId('loading-spinner')).toBeInTheDocument();

    await waitFor(() => {
      expect(mockRenewCertificate).toHaveBeenCalled();
    });
  });

  it('should display serial numbers correctly', () => {
    render(<CertificateTable />);

    expect(screen.getByText('123456789')).toBeInTheDocument();
    expect(screen.getByText('987654321')).toBeInTheDocument();
    expect(screen.getByText('555666777')).toBeInTheDocument();
  });

  it('should render certificate details links', () => {
    render(<CertificateTable />);

    // Check that certificate names are rendered as links
    const certificateLinks = screen.getAllByRole('link');
    expect(certificateLinks).toHaveLength(3);
    
    // Check specific certificate link
    expect(screen.getByRole('link', { name: 'example.com' })).toHaveAttribute(
      'href',
      '/certificates/123456789'
    );
  });
}); 