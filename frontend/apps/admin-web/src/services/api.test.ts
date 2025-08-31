import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ApiService } from './api'

// Mock fetch
const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

describe('ApiService', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })

  it('healthCheck returns parsed JSON response', async () => {
    const mockResponse = {
      status: 'ok',
      message: 'Multi-tenant admin backend is running',
      database: 'healthy',
      redis: 'healthy'
    }

    const mockFetch = vi.mocked(fetch)
    mockFetch.mockResolvedValueOnce({
      json: async () => mockResponse,
    } as Response)

    const result = await ApiService.healthCheck()

    expect(mockFetch).toHaveBeenCalledWith('http://localhost:8000/health')
    expect(result).toEqual(mockResponse)
  })

  it('healthCheck throws error on fetch failure', async () => {
    const mockFetch = vi.mocked(fetch)
    mockFetch.mockRejectedValueOnce(new Error('Network error'))

    await expect(ApiService.healthCheck()).rejects.toThrow('Network error')
  })

  it('get method makes correct API call', async () => {
    const mockResponse = { data: 'test' }
    
    const mockFetch = vi.mocked(fetch)
    mockFetch.mockResolvedValueOnce({
      json: async () => mockResponse,
    } as Response)

    const result = await ApiService.get('/test')

    expect(mockFetch).toHaveBeenCalledWith('http://localhost:8000/test')
    expect(result).toEqual(mockResponse)
  })
})