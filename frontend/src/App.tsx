import { useEffect } from 'react'
import Game from './game/Game'
import './App.css'

function App() {
  // Fetch high score or initial state from backend
  useEffect(() => {
    fetch('/api/score')
      .then(res => res.json())
      .then(data => console.log("High Score from backend:", data.highScore))
      .catch(err => console.error(err))
  }, [])

  return (
    <>
      <header>
        <h1>*** PACMAN C64 ***</h1>
        <p>READY.</p>
      </header>
      <main>
        <div className="game-container">
           <Game />
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
