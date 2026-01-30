import { useEffect, useState } from 'react'
import Game from './game/Game'
import AuthForm from './components/AuthForm'
import './App.css'

const USERNAME_STORAGE_KEY = 'pacman.username'

function App() {
  const [username, setUsername] = useState<string | null>(null)

  // Fetch high score or initial state from backend
  useEffect(() => {
    fetch('/api/score')
      .then(res => res.json())
      .then(data => console.log("High Score from backend:", data.highScore))
      .catch(err => console.error(err))
  }, [])

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

  return (
    <>
      <header>
        <h1>*** PACMAN C64 ***</h1>
        <p>{username ? 'READY.' : 'AUTHENTICATION REQUIRED.'}</p>
        {username && <p>PLAYER: {username}</p>}
      </header>
      <main>
        <div className="game-container">
          {username ? (
            <Game />
          ) : (
            <AuthForm onLoginSuccess={handleLoginSuccess} />
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
