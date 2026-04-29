import React from 'react'
import {createRoot} from 'react-dom/client'
import './style.css'
import App from './App'

class RenderBoundary extends React.Component<{ children: React.ReactNode }, { error: Error | null }> {
    constructor(props: { children: React.ReactNode }) {
        super(props)
        this.state = {error: null}
    }

    static getDerivedStateFromError(error: Error) {
        return {error}
    }

    render() {
        if (!this.state.error) {
            return this.props.children
        }

        return (
            <main className="boot-screen">
                <div className="boot-card">
                    <div className="brand-mark">NX</div>
                    <h1>NEXUS</h1>
                    <p>{this.state.error.message}</p>
                    <button type="button" onClick={() => window.location.reload()}>Retry</button>
                </div>
            </main>
        )
    }
}

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <RenderBoundary>
            <App/>
        </RenderBoundary>
    </React.StrictMode>
)
