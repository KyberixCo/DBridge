import React from 'react'
import {createRoot} from 'react-dom/client'
import './style.css'
import App from './App'
import { LanguageProvider } from './i18n/LanguageContext'

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <LanguageProvider>
            <App/>
        </LanguageProvider>
    </React.StrictMode>
)
