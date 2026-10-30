import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import App from './App'
import { AppDataProvider } from './data'
import { RewardsProvider } from './rewards'
import './index.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <AppDataProvider>
        <RewardsProvider>
          <App />
        </RewardsProvider>
      </AppDataProvider>
    </BrowserRouter>
  </StrictMode>,
)
