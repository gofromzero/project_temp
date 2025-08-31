import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import App from './App'

describe('App', () => {
  it('renders Hello World message', () => {
    render(<App />)
    
    // Check for key elements
    expect(screen.getByText('Hello World!')).toBeInTheDocument()
    expect(screen.getByText('Multi-Tenant Admin')).toBeInTheDocument()
    expect(screen.getByText('Welcome to the Multi-Tenant Admin System')).toBeInTheDocument()
    expect(screen.getByText('Test Backend Connection')).toBeInTheDocument()
  })

  it('renders technology stack information', () => {
    render(<App />)
    
    expect(screen.getByText(/Frontend: React \+ TypeScript \+ Ant Design \+ Tailwind CSS/)).toBeInTheDocument()
  })

  it('shows test backend connection button', () => {
    render(<App />)
    
    const button = screen.getByRole('button', { name: /Test Backend Connection/ })
    expect(button).toBeInTheDocument()
  })
})