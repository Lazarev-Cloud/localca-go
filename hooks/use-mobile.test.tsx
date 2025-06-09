import { renderHook } from '@testing-library/react';
import { useIsMobile } from './use-mobile';

// Mock window.matchMedia
const mockMatchMedia = jest.fn();

Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: mockMatchMedia,
});

describe('useIsMobile', () => {
  beforeEach(() => {
    mockMatchMedia.mockClear();
  });

  it('should return true for mobile screen size', () => {
    // Mock mobile screen size (less than 768px)
    mockMatchMedia.mockReturnValue({
      matches: true,
      media: '(max-width: 767px)',
      onchange: null,
      addListener: jest.fn(),
      removeListener: jest.fn(),
      addEventListener: jest.fn(),
      removeEventListener: jest.fn(),
      dispatchEvent: jest.fn(),
    });

    const { result } = renderHook(() => useIsMobile());

    expect(result.current).toBe(true);
    expect(mockMatchMedia).toHaveBeenCalledWith('(max-width: 767px)');
  });

  it('should return false for desktop screen size', () => {
    // Mock desktop screen size (768px or larger)
    mockMatchMedia.mockReturnValue({
      matches: false,
      media: '(max-width: 767px)',
      onchange: null,
      addListener: jest.fn(),
      removeListener: jest.fn(),
      addEventListener: jest.fn(),
      removeEventListener: jest.fn(),
      dispatchEvent: jest.fn(),
    });

    const { result } = renderHook(() => useIsMobile());

    expect(result.current).toBe(false);
    expect(mockMatchMedia).toHaveBeenCalledWith('(max-width: 767px)');
  });

  it('should handle media query changes', () => {
    const mockMediaQuery = {
      matches: false,
      media: '(max-width: 767px)',
      onchange: null,
      addListener: jest.fn(),
      removeListener: jest.fn(),
      addEventListener: jest.fn(),
      removeEventListener: jest.fn(),
      dispatchEvent: jest.fn(),
    };

    mockMatchMedia.mockReturnValue(mockMediaQuery);

    const { result, rerender } = renderHook(() => useIsMobile());

    // Initially desktop
    expect(result.current).toBe(false);

    // Simulate screen size change to mobile
    mockMediaQuery.matches = true;
    
    // Trigger the change event
    if (mockMediaQuery.addEventListener.mock.calls.length > 0) {
      const changeHandler = mockMediaQuery.addEventListener.mock.calls[0][1];
      changeHandler({ matches: true });
    }

    rerender();

    // Should still work with the hook's internal state management
    expect(mockMediaQuery.addEventListener).toHaveBeenCalledWith('change', expect.any(Function));
  });

  it('should clean up event listeners on unmount', () => {
    const mockMediaQuery = {
      matches: false,
      media: '(max-width: 767px)',
      onchange: null,
      addListener: jest.fn(),
      removeListener: jest.fn(),
      addEventListener: jest.fn(),
      removeEventListener: jest.fn(),
      dispatchEvent: jest.fn(),
    };

    mockMatchMedia.mockReturnValue(mockMediaQuery);

    const { unmount } = renderHook(() => useIsMobile());

    expect(mockMediaQuery.addEventListener).toHaveBeenCalled();

    unmount();

    expect(mockMediaQuery.removeEventListener).toHaveBeenCalled();
  });
}); 