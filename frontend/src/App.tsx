import { useEffect, useState, type FormEvent } from 'react'
import Game from './game/Game'
import './App.css'

const USERNAME_STORAGE_KEY = 'pacman.username'

function App() {
  const [username, setUsername] = useState<string | null>(null)
  const [usernameInput, setUsernameInput] = useState('')
  const [usernameError, setUsernameError] = useState('')

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

  const handleUsernameSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const trimmed = usernameInput.trim()
    if (!trimmed) {
      setUsernameError('Please enter a username.')
      return
    }
    sessionStorage.setItem(USERNAME_STORAGE_KEY, trimmed)
    setUsername(trimmed)
    setUsernameError('')
  }

  return (
    <>
      <header>
        <h1>*** PACMAN C64 ***</h1>
        <p>{username ? 'READY.' : 'ENTER USERNAME TO START.'}</p>
        {username && <p>PLAYER: {username}</p>}
      </header>
      <main>
        {!username && (
          <div className="username-section" role="dialog" aria-labelledby="username-title">
            <form className="username-form" onSubmit={handleUsernameSubmit}>
              <h2 id="username-title">PLAYER NAME</h2>
              <label htmlFor="username-input">USERNAME</label>
              <input
                id="username-input"
                type="text"
                value={usernameInput}
                onChange={event => setUsernameInput(event.target.value)}
                autoFocus
              />
              {usernameError && <p className="username-error">{usernameError}</p>}
              <button type="submit">START</button>
            </form>
          </div>
        )}
        <div className="game-container">
          {username && <Game />}
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
