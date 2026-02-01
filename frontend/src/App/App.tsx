import { useState } from 'react'
import Game from '../game/Game/Game'
import type { GameMode } from '../game/constants'
import AuthForm from '../components/AuthForm/AuthForm'
import ScoreBoard from '../components/ScoreBoard'
import './App.css'

const USERNAME_STORAGE_KEY = 'pacman.username'
const TOKEN_STORAGE_KEY = 'pacman.token'

type ViewState = 'game' | 'scoreboard'

function App() {
  const [username, setUsername] = useState<string | null>(() => {
    const stored = sessionStorage.getItem(USERNAME_STORAGE_KEY)
    const trimmed = stored?.trim()
    return trimmed || null
  })
  const [authToken, setAuthToken] = useState<string | null>(() => {
    return sessionStorage.getItem(TOKEN_STORAGE_KEY)
  })
  
  const [currentView, setCurrentView] = useState<ViewState>('game')
  const [onlineCount, setOnlineCount] = useState<number>(0)
  const [ghostCount, setGhostCount] = useState<number>(4)

  const handleLoginSuccess = (nickname: string, token: string) => {
    sessionStorage.setItem(USERNAME_STORAGE_KEY, nickname)
    sessionStorage.setItem(TOKEN_STORAGE_KEY, token)
    setUsername(nickname)
    setAuthToken(token)
  }

  const handleLogout = () => {
    sessionStorage.removeItem(USERNAME_STORAGE_KEY)
    sessionStorage.removeItem(TOKEN_STORAGE_KEY)
    setUsername(null)
    setAuthToken(null)
    setCurrentView('game')
    setOnlineCount(0)
  }

  const [scoreboardMode, setScoreboardMode] = useState<GameMode>('single')

  const handleShowScoreboard = (count?: number, mode?: GameMode) => {
    if (count) {
      setGhostCount(count)
    }
    if (mode) {
      setScoreboardMode(mode)
    }
    setCurrentView('scoreboard')
  }

  const handleBackToGame = () => {
    setCurrentView('game')
  }

  const isAuthenticated = username && authToken

  return (
    <>
      <header>
        <h1>*** PACMAN C64 ***</h1>
        <p>{isAuthenticated ? 'READY.' : 'AUTHENTICATION REQUIRED.'}</p>
        {isAuthenticated && <p>PLAYER: {username}</p>}
        {isAuthenticated && <p>ONLINE: {onlineCount}</p>}
      </header>
      <main>
        <div className="game-container">
          {!isAuthenticated ? (
            <AuthForm onLoginSuccess={handleLoginSuccess} />
          ) : currentView === 'scoreboard' ? (
            <ScoreBoard onBack={handleBackToGame} initialGhostCount={ghostCount} activeMode={scoreboardMode} />
          ) : (
            <Game 
              onLogout={handleLogout}
              onShowScoreboard={handleShowScoreboard}
              onOnlineCountChange={setOnlineCount}
              username={username}
              authToken={authToken}
              ghostCount={ghostCount}
              onGhostCountChange={setGhostCount}
            />
          )}
        </div>
      </main>
      <footer>
        <p>COMMODORE 64 BASIC V2</p>
        <p>64K RAM SYSTEM  38911 BASIC BYTES FREE</p>
      </footer>
    </>
  )
}

export default App
