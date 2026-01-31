import { useEffect, useState } from 'react'
import Game from './game/Game'
import AuthForm from './components/AuthForm'
import ScoreBoard from './components/ScoreBoard'
import './App.css'

const USERNAME_STORAGE_KEY = 'pacman.username'

type ViewState = 'game' | 'scoreboard'

function App() {
  const [username, setUsername] = useState<string | null>(null)
  const [currentView, setCurrentView] = useState<ViewState>('game')

  useEffect(() => {
    const stored = sessionStorage.getItem(USERNAME_STORAGE_KEY)
    const trimmed = stored?.trim()
    if (trimmed) {
      setUsername(trimmed)
    }
  }, [])

  const handleLoginSuccess = (nickname: string) => {
    sessionStorage.setItem(USERNAME_STORAGE_KEY, nickname)
    setUsername(nickname)
  }

  const handleLogout = () => {
    sessionStorage.removeItem(USERNAME_STORAGE_KEY)
    setUsername(null)
    setCurrentView('game')
  }

  const handleShowScoreboard = () => {
    setCurrentView('scoreboard')
  }

  const handleBackToGame = () => {
    setCurrentView('game')
  }

  return (
    <>
      <header>
        <h1>*** PACMAN C64 ***</h1>
        <p>{username ? 'READY.' : 'AUTHENTICATION REQUIRED.'}</p>
        {username && <p>PLAYER: {username}</p>}
      </header>
      <main>
        <div className="game-container">
          {!username ? (
            <AuthForm onLoginSuccess={handleLoginSuccess} />
          ) : currentView === 'scoreboard' ? (
            <ScoreBoard onBack={handleBackToGame} />
          ) : (
            <Game 
              onLogout={handleLogout}
              onShowScoreboard={handleShowScoreboard}
              username={username}
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

