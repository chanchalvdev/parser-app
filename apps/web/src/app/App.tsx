import { AppRouter } from './router'
import { AppShell } from '@/components/layout/AppShell'

export const App = () => {
  return (
    <AppShell>
      <AppRouter />
    </AppShell>
  )
}

export default App
