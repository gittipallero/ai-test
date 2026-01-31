import { useState, type FormEvent } from 'react'
import './AuthForm.css'

interface AuthResponse {
  message: string
  nickname: string
  token: string
}

interface AuthFormProps {
  onLoginSuccess: (nickname: string, token: string) => void
}

export default function AuthForm({ onLoginSuccess }: AuthFormProps) {
  const [isLoginMode, setIsLoginMode] = useState(true)
  const [authNickname, setAuthNickname] = useState('')
  const [authPassword, setAuthPassword] = useState('')
  const [authError, setAuthError] = useState('')
  const [authSuccess, setAuthSuccess] = useState('')

  const handleAuthSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setAuthError('')
    setAuthSuccess('')

    if (!authNickname.trim() || !authPassword.trim()) {
      setAuthError('Nickname and Password are required.')
      return
    }

    const endpoint = isLoginMode ? '/api/login' : '/api/signup'

    try {
      const response = await fetch(endpoint, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ nickname: authNickname, password: authPassword })
      })

      if (!response.ok) {
        if (response.status === 409) throw new Error('Nickname already exists.')
        if (response.status === 401) throw new Error('Invalid credentials.')
        const errText = await response.text() 
        throw new Error(errText || 'An error occurred.')
      }

      const data: AuthResponse = await response.json()
      
      // Both login and signup now return a token (auto-login after signup)
      onLoginSuccess(data.nickname, data.token)
    } catch (err) {
      if (err instanceof Error) {
        setAuthError(err.message)
      } else {
        setAuthError('An unexpected error occurred.')
      }
    }
  }

  const toggleAuthMode = () => {
    setIsLoginMode(!isLoginMode)
    setAuthError('')
    setAuthSuccess('')
    setAuthPassword('')
  }

  return (
    <div className="username-overlay" role="dialog" aria-modal="true" aria-labelledby="auth-title">
      <form className="username-form" onSubmit={handleAuthSubmit}>
        <h2 id="auth-title">{isLoginMode ? 'LOGIN' : 'SIGNUP'}</h2>
        
        <label htmlFor="auth-nickname">NICKNAME</label>
        <input
          id="auth-nickname"
          type="text"
          value={authNickname}
          onChange={e => setAuthNickname(e.target.value)}
          autoFocus
        />

        <label htmlFor="auth-password">PASSWORD</label>
        <input
          id="auth-password"
          type="password"
          value={authPassword}
          onChange={e => setAuthPassword(e.target.value)}
        />

        {authError && <p className="username-error">{authError}</p>}
        {authSuccess && <p className="username-success" style={{color: 'lightgreen'}}>{authSuccess}</p>}

        <button type="submit">{isLoginMode ? 'LOGIN' : 'SIGN UP'}</button>
        
        <p style={{marginTop: '1rem', fontSize: '0.8rem', cursor: 'pointer', textDecoration: 'underline'}} onClick={toggleAuthMode}>
          {isLoginMode ? 'NEED A ACCOUNT? SIGN UP' : 'ALREADY HAVE ACCOUNT? LOGIN'}
        </p>
      </form>
    </div>
  )
}
