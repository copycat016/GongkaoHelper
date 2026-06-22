import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.jsx'
import { getThemeConfig } from './api/theme.js'
import { applyThemeConfig } from './theme/applyThemeConfig.js'

async function init() {
  const config = await getThemeConfig()
  applyThemeConfig(config)
  
  createRoot(document.getElementById('root')).render(
    <StrictMode>
      <App initialPalette={config.palette} />
    </StrictMode>,
  )
}

init()
