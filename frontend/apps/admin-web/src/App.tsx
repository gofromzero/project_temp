import { ConfigProvider, Layout, Typography, Card, Button, message } from 'antd'
import { RocketOutlined, ApiOutlined } from '@ant-design/icons'
import { useState } from 'react'
import { ApiService } from './services/api'
import './styles/App.css'

const { Header, Content } = Layout
const { Title, Text } = Typography

function App() {
  const [backendStatus, setBackendStatus] = useState<any>(null)
  const [loading, setLoading] = useState(false)

  const testBackendConnection = async () => {
    setLoading(true)
    try {
      const response = await ApiService.healthCheck()
      setBackendStatus(response)
      message.success('Backend connection successful!')
    } catch (error) {
      console.error('Backend connection failed:', error)
      message.error('Backend connection failed. Make sure the backend server is running on port 8000.')
      setBackendStatus({ status: 'error', message: 'Connection failed' })
    } finally {
      setLoading(false)
    }
  }

  return (
    <ConfigProvider
      theme={{
        token: {
          colorPrimary: '#1677ff',
        },
      }}
    >
      <Layout className="min-h-screen">
        <Header className="bg-white shadow-md">
          <div className="flex items-center">
            <RocketOutlined className="text-xl mr-2" />
            <Title level={3} className="m-0">
              Multi-Tenant Admin
            </Title>
          </div>
        </Header>
        <Content className="p-6">
          <div className="max-w-4xl mx-auto">
            <Card className="text-center">
              <RocketOutlined className="text-6xl text-blue-500 mb-4" />
              <Title level={1} className="mb-4">
                Hello World!
              </Title>
              <Text className="text-lg">
                Welcome to the Multi-Tenant Admin System
              </Text>
              
              <div className="mt-6 p-4 bg-gray-50 rounded">
                <Text type="secondary">
                  Frontend: React + TypeScript + Ant Design + Tailwind CSS
                </Text>
              </div>

              <div className="mt-6">
                <Button 
                  type="primary" 
                  icon={<ApiOutlined />}
                  loading={loading}
                  onClick={testBackendConnection}
                  size="large"
                >
                  Test Backend Connection
                </Button>
                
                {backendStatus && (
                  <div className="mt-4 p-4 border rounded">
                    <Text strong>Backend Status:</Text>
                    <pre className="mt-2 text-left bg-gray-100 p-2 rounded text-sm">
                      {JSON.stringify(backendStatus, null, 2)}
                    </pre>
                  </div>
                )}
              </div>
            </Card>
          </div>
        </Content>
      </Layout>
    </ConfigProvider>
  )
}

export default App