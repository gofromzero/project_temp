const API_BASE_URL = 'http://localhost:8000'

export class ApiService {
  static async healthCheck() {
    try {
      const response = await fetch(`${API_BASE_URL}/health`)
      return await response.json()
    } catch (error) {
      console.error('Health check failed:', error)
      throw error
    }
  }

  static async get(endpoint: string) {
    try {
      const response = await fetch(`${API_BASE_URL}${endpoint}`)
      return await response.json()
    } catch (error) {
      console.error(`API GET ${endpoint} failed:`, error)
      throw error
    }
  }
}