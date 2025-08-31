import { Layout } from 'antd'
import { ReactNode } from 'react'

const { Header, Content } = Layout

interface AppLayoutProps {
  children: ReactNode
}

export const AppLayout = ({ children }: AppLayoutProps) => {
  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header>
        <div style={{ color: 'white' }}>Multi-Tenant Admin</div>
      </Header>
      <Layout>
        <Content style={{ padding: '24px' }}>
          {children}
        </Content>
      </Layout>
    </Layout>
  )
}