# LocalCA Development Guide

This guide covers everything you need to know for developing LocalCA, from initial setup to advanced development workflows.

## Development Environment Setup

### Prerequisites

- **Go 1.23+**: Backend development
- **Node.js 18+**: Frontend development
- **Docker & Docker Compose**: Containerized development
- **Git**: Version control
- **VS Code or similar**: IDE with Go and TypeScript support

### Initial Setup

1. **Clone the repository**:
```bash
git clone https://github.com/Lazarev-Cloud/localca-go.git
cd localca-go
```

2. **Install backend dependencies**:
```bash
go mod download
```

3. **Install frontend dependencies**:
```bash
npm install
```

4. **Create development environment file**:
```bash
cp .env.example .env.dev
```

Edit `.env.dev` for development:
```bash
# Development Configuration
CA_NAME=LocalCA-Dev
CA_KEY_PASSWORD=dev-password
ORGANIZATION=LocalCA Development
COUNTRY=US

# Development URLs
LISTEN_ADDR=:8080
NEXT_PUBLIC_API_URL=http://localhost:8080

# Enhanced Storage (optional for development)
DATABASE_ENABLED=false
S3_ENABLED=false
CACHE_ENABLED=false

# Logging
LOG_FORMAT=text
LOG_LEVEL=debug

# Development Security (less strict)
TLS_ENABLED=false
SESSION_SECRET=dev-session-secret
```

## Development Workflow

### Running the Application

#### Option 1: Docker Compose (Recommended)
```bash
# Start all services
docker-compose -f docker-compose.dev.yml up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

#### Option 2: Manual Development
```bash
# Terminal 1: Start backend
go run main.go

# Terminal 2: Start frontend
npm run dev

# Terminal 3: Start enhanced storage (optional)
docker-compose up postgres minio keydb
```

#### Option 3: Development Scripts
```bash
# Use development scripts
./run-dev.sh    # Linux/macOS
./run-dev.bat   # Windows
```

### Development URLs

- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **API Documentation**: http://localhost:8080/api/docs (if enabled)
- **MinIO Console**: http://localhost:9001 (if using enhanced storage)

## Project Architecture

### Backend Structure

```
pkg/
├── acme/           # ACME protocol implementation
├── cache/          # Redis/KeyDB caching layer
├── certificates/   # Certificate management core
├── config/         # Configuration management
├── database/       # PostgreSQL integration
├── email/          # Email notification service
├── handlers/       # HTTP handlers and routing
├── logging/        # Structured logging
├── s3storage/      # S3/MinIO object storage
├── security/       # Security utilities
└── storage/        # Storage interfaces and implementations
```

### Frontend Structure

```
app/
├── api/            # Next.js API routes
├── certificates/   # Certificate management pages
├── create/         # Certificate creation flow
├── login/          # Authentication pages
├── settings/       # Application settings
└── setup/          # Initial setup flow

components/
├── ui/             # Base UI components (shadcn/ui)
├── forms/          # Form components
├── layout/         # Layout components
└── certificates/   # Certificate-specific components

hooks/
├── use-api.ts      # Generic API client
├── use-auth.ts     # Authentication management
└── use-*.ts        # Feature-specific hooks
```

## Development Standards

### Go Backend Standards

#### Code Organization
```go
// Package structure example
package certificates

import (
    "context"
    "crypto/x509"
    "time"
    
    "github.com/Lazarev-Cloud/localca-go/pkg/config"
    "github.com/Lazarev-Cloud/localca-go/pkg/storage"
)

// Service interface
type CertificateService interface {
    CreateServerCertificate(ctx context.Context, req *CreateServerCertRequest) (*Certificate, error)
    CreateClientCertificate(ctx context.Context, req *CreateClientCertRequest) (*Certificate, error)
    RevokeCertificate(ctx context.Context, id string, reason RevocationReason) error
}

// Implementation
type certificateService struct {
    config  *config.Config
    storage storage.StorageInterface
    logger  *logrus.Logger
}
```

#### Error Handling
```go
// Custom error types
type CertificateError struct {
    Type    string
    Message string
    Cause   error
}

func (e *CertificateError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
    }
    return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Error handling in functions
func (s *certificateService) CreateCertificate(req *CreateCertRequest) (*Certificate, error) {
    if err := validateRequest(req); err != nil {
        return nil, &CertificateError{
            Type:    "ValidationError",
            Message: "Invalid certificate request",
            Cause:   err,
        }
    }
    
    cert, err := s.generateCertificate(req)
    if err != nil {
        s.logger.WithError(err).Error("Failed to generate certificate")
        return nil, &CertificateError{
            Type:    "GenerationError",
            Message: "Failed to generate certificate",
            Cause:   err,
        }
    }
    
    return cert, nil
}
```

#### Testing
```go
// Test structure
func TestCertificateService_CreateServerCertificate(t *testing.T) {
    tests := []struct {
        name    string
        request *CreateServerCertRequest
        want    *Certificate
        wantErr bool
    }{
        {
            name: "valid server certificate",
            request: &CreateServerCertRequest{
                CommonName: "example.com",
                SANs:       []string{"www.example.com"},
                ValidDays:  365,
            },
            wantErr: false,
        },
        {
            name: "invalid common name",
            request: &CreateServerCertRequest{
                CommonName: "",
                ValidDays:  365,
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := setupTestService(t)
            got, err := service.CreateServerCertificate(context.Background(), tt.request)
            
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.NotNil(t, got)
            assert.Equal(t, tt.request.CommonName, got.CommonName)
        })
    }
}
```

### Frontend Standards

#### Component Structure
```typescript
// Component with TypeScript
interface CertificateCardProps {
  certificate: Certificate;
  onRenew?: (id: string) => void;
  onRevoke?: (id: string) => void;
  onDelete?: (id: string) => void;
}

export function CertificateCard({ 
  certificate, 
  onRenew, 
  onRevoke, 
  onDelete 
}: CertificateCardProps) {
  const [isLoading, setIsLoading] = useState(false);
  
  const handleRenew = async () => {
    if (!onRenew) return;
    
    setIsLoading(true);
    try {
      await onRenew(certificate.id);
      toast.success('Certificate renewed successfully');
    } catch (error) {
      toast.error('Failed to renew certificate');
    } finally {
      setIsLoading(false);
    }
  };
  
  return (
    <Card className="p-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="font-semibold">{certificate.commonName}</h3>
          <p className="text-sm text-muted-foreground">
            Expires: {format(new Date(certificate.validTo), 'PPP')}
          </p>
        </div>
        <div className="flex gap-2">
          <Button 
            variant="outline" 
            size="sm" 
            onClick={handleRenew}
            disabled={isLoading}
          >
            {isLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Renew'}
          </Button>
        </div>
      </div>
    </Card>
  );
}
```

#### Custom Hooks
```typescript
// API integration hook
export function useCertificates() {
  const [certificates, setCertificates] = useState<Certificate[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  const fetchCertificates = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await fetch('/api/proxy/certificates', {
        credentials: 'include'
      });
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      
      const data = await response.json();
      setCertificates(data.certificates || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  }, []);
  
  useEffect(() => {
    fetchCertificates();
  }, [fetchCertificates]);
  
  const createCertificate = useCallback(async (request: CreateCertificateRequest) => {
    const response = await fetch('/api/proxy/certificates', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify(request)
    });
    
    if (!response.ok) {
      throw new Error('Failed to create certificate');
    }
    
    await fetchCertificates(); // Refresh list
    return response.json();
  }, [fetchCertificates]);
  
  return {
    certificates,
    loading,
    error,
    createCertificate,
    refetch: fetchCertificates
  };
}
```

#### Form Validation
```typescript
// Zod schema for validation
const createCertificateSchema = z.object({
  type: z.enum(['server', 'client']),
  commonName: z.string().min(1, 'Common name is required'),
  subjectAltNames: z.array(z.string()).optional(),
  organization: z.string().optional(),
  country: z.string().length(2, 'Country must be 2 characters').optional(),
  validityDays: z.number().min(1).max(3650),
  keySize: z.enum([2048, 4096]),
  keyType: z.enum(['RSA', 'ECDSA'])
});

type CreateCertificateForm = z.infer<typeof createCertificateSchema>;

// Form component
export function CreateCertificateForm() {
  const form = useForm<CreateCertificateForm>({
    resolver: zodResolver(createCertificateSchema),
    defaultValues: {
      type: 'server',
      validityDays: 365,
      keySize: 2048,
      keyType: 'RSA'
    }
  });
  
  const onSubmit = async (data: CreateCertificateForm) => {
    try {
      await createCertificate(data);
      toast.success('Certificate created successfully');
      form.reset();
    } catch (error) {
      toast.error('Failed to create certificate');
    }
  };
  
  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="commonName"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Common Name</FormLabel>
              <FormControl>
                <Input placeholder="example.com" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        {/* Additional form fields */}
      </form>
    </Form>
  );
}
```

## Testing Strategy

### Backend Testing

#### Unit Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./pkg/certificates
```

#### Integration Tests
```bash
# Run integration tests
go test -tags=integration ./...

# Run with test database
TEST_DATABASE_URL=postgres://test:test@localhost:5433/test_localca go test ./...
```

#### Test Structure
```go
// Test setup
func setupTestService(t *testing.T) *certificateService {
    config := &config.Config{
        CAName:       "Test CA",
        Organization: "Test Org",
        Country:      "US",
    }
    
    storage := &mockStorage{}
    logger := logrus.New()
    logger.SetLevel(logrus.DebugLevel)
    
    return &certificateService{
        config:  config,
        storage: storage,
        logger:  logger,
    }
}

// Mock storage
type mockStorage struct {
    certificates map[string]*Certificate
}

func (m *mockStorage) StoreCertificate(cert *Certificate) error {
    if m.certificates == nil {
        m.certificates = make(map[string]*Certificate)
    }
    m.certificates[cert.ID] = cert
    return nil
}
```

### Frontend Testing

#### Component Tests
```bash
# Run frontend tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage
```

#### Test Structure
```typescript
// Component test
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { CertificateCard } from './certificate-card';

const mockCertificate: Certificate = {
  id: 'cert-123',
  commonName: 'example.com',
  type: 'server',
  status: 'valid',
  validFrom: '2024-01-01T00:00:00Z',
  validTo: '2025-01-01T00:00:00Z',
  serialNumber: '123456789'
};

describe('CertificateCard', () => {
  it('renders certificate information', () => {
    render(<CertificateCard certificate={mockCertificate} />);
    
    expect(screen.getByText('example.com')).toBeInTheDocument();
    expect(screen.getByText(/Expires:/)).toBeInTheDocument();
  });
  
  it('calls onRenew when renew button is clicked', async () => {
    const onRenew = jest.fn();
    render(<CertificateCard certificate={mockCertificate} onRenew={onRenew} />);
    
    fireEvent.click(screen.getByText('Renew'));
    
    await waitFor(() => {
      expect(onRenew).toHaveBeenCalledWith('cert-123');
    });
  });
});
```

#### API Mocking
```typescript
// Mock API responses
import { rest } from 'msw';
import { setupServer } from 'msw/node';

const server = setupServer(
  rest.get('/api/proxy/certificates', (req, res, ctx) => {
    return res(
      ctx.json({
        certificates: [mockCertificate],
        total: 1
      })
    );
  }),
  
  rest.post('/api/proxy/certificates', (req, res, ctx) => {
    return res(
      ctx.json({
        success: true,
        certificate: mockCertificate
      })
    );
  })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());
```

## Debugging

### Backend Debugging

#### VS Code Configuration
```json
// .vscode/launch.json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug LocalCA Backend",
      "type": "go",
      "request": "launch",
      "mode": "auto",
      "program": "${workspaceFolder}/main.go",
      "env": {
        "CA_KEY_PASSWORD": "dev-password",
        "LOG_LEVEL": "debug"
      },
      "args": []
    }
  ]
}
```

#### Logging
```go
// Structured logging
logger.WithFields(logrus.Fields{
    "certificate_id": certID,
    "common_name":    commonName,
    "operation":      "create",
}).Info("Creating certificate")

// Error logging with context
logger.WithError(err).WithFields(logrus.Fields{
    "certificate_id": certID,
    "step":          "key_generation",
}).Error("Failed to generate private key")
```

### Frontend Debugging

#### Browser DevTools
- Use React Developer Tools extension
- Enable source maps for debugging TypeScript
- Use Network tab to debug API calls

#### Debug Configuration
```typescript
// Debug API calls
const DEBUG = process.env.NODE_ENV === 'development';

export async function apiCall(endpoint: string, options?: RequestInit) {
  if (DEBUG) {
    console.log(`API Call: ${endpoint}`, options);
  }
  
  const response = await fetch(endpoint, options);
  
  if (DEBUG) {
    console.log(`API Response: ${endpoint}`, {
      status: response.status,
      headers: Object.fromEntries(response.headers.entries())
    });
  }
  
  return response;
}
```

## Performance Optimization

### Backend Performance

#### Database Optimization
```go
// Connection pooling
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)

// Query optimization
func (s *storage) GetCertificatesByStatus(status string) ([]*Certificate, error) {
    query := `
        SELECT id, common_name, type, status, valid_from, valid_to 
        FROM certificates 
        WHERE status = $1 
        ORDER BY created_at DESC
    `
    
    rows, err := s.db.Query(query, status)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var certificates []*Certificate
    for rows.Next() {
        cert := &Certificate{}
        err := rows.Scan(&cert.ID, &cert.CommonName, &cert.Type, 
                        &cert.Status, &cert.ValidFrom, &cert.ValidTo)
        if err != nil {
            return nil, err
        }
        certificates = append(certificates, cert)
    }
    
    return certificates, nil
}
```

#### Caching
```go
// Cache implementation
func (s *cachedStorage) GetCertificate(id string) (*Certificate, error) {
    // Check cache first
    if cert, found := s.cache.Get(fmt.Sprintf("cert:%s", id)); found {
        return cert.(*Certificate), nil
    }
    
    // Fetch from storage
    cert, err := s.storage.GetCertificate(id)
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    s.cache.Set(fmt.Sprintf("cert:%s", id), cert, time.Hour)
    
    return cert, nil
}
```

### Frontend Performance

#### Code Splitting
```typescript
// Lazy loading components
const CertificateDetails = lazy(() => import('./certificate-details'));
const CreateCertificate = lazy(() => import('./create-certificate'));

// Route-based code splitting
export default function App() {
  return (
    <Router>
      <Suspense fallback={<Loading />}>
        <Routes>
          <Route path="/certificates/:id" element={<CertificateDetails />} />
          <Route path="/create" element={<CreateCertificate />} />
        </Routes>
      </Suspense>
    </Router>
  );
}
```

#### Memoization
```typescript
// Memoized components
const CertificateCard = memo(({ certificate, onRenew }: CertificateCardProps) => {
  const handleRenew = useCallback(() => {
    onRenew?.(certificate.id);
  }, [certificate.id, onRenew]);
  
  return (
    <Card>
      {/* Component content */}
    </Card>
  );
});

// Memoized values
const filteredCertificates = useMemo(() => {
  return certificates.filter(cert => 
    cert.commonName.toLowerCase().includes(searchTerm.toLowerCase())
  );
}, [certificates, searchTerm]);
```

## Contributing Guidelines

### Git Workflow

1. **Create feature branch**:
```bash
git checkout -b feature/certificate-renewal
```

2. **Make changes and commit**:
```bash
git add .
git commit -m "feat: add automatic certificate renewal"
```

3. **Push and create PR**:
```bash
git push origin feature/certificate-renewal
# Create pull request on GitHub
```

### Commit Message Format

Follow conventional commits:
```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance tasks

### Code Review Checklist

- [ ] Code follows project standards
- [ ] Tests are included and passing
- [ ] Documentation is updated
- [ ] No security vulnerabilities
- [ ] Performance impact considered
- [ ] Backward compatibility maintained

## Deployment

### Development Deployment
```bash
# Build and test
make build
make test

# Deploy to development environment
make deploy-dev
```

### Production Deployment
```bash
# Build production images
docker build -t localca/backend:latest .
docker build -f Dockerfile.frontend -t localca/frontend:latest .

# Deploy to production
docker-compose -f docker-compose.prod.yml up -d
```

This development guide provides comprehensive information for contributing to LocalCA. Follow these guidelines to ensure consistent, high-quality code and smooth collaboration. 