import { useState } from 'react'
import Game from './game/Game'
import AuthForm from './components/AuthForm'
import ScoreBoard from './components/ScoreBoard'
import './App.css'

const USERNAME_STORAGE_KEY = 'pacman.username'

type ViewState = 'game' | 'scoreboard'

function App() {
  const [username, setUsername] = useState<string | null>(() => {
    const stored = sessionStorage.getItem(USERNAME_STORAGE_KEY)
    const trimmed = stored?.trim()
    return trimmed || null
  })
  const [currentView, setCurrentView] = useState<ViewState>('game')
  const [onlineCount, setOnlineCount] = useState<number>(0)

  const handleLoginSuccess = (nickname: string) => {
    sessionStorage.setItem(USERNAME_STORAGE_KEY, nickname)
    setUsername(nickname)
  }

  const handleLogout = () => {
    sessionStorage.removeItem(USERNAME_STORAGE_KEY)
    setUsername(null)
    setCurrentView('game')
    setOnlineCount(0)
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
        {username && <p>ONLINE: {onlineCount}</p>}
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
              onOnlineCountChange={setOnlineCount}
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

