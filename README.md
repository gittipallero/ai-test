# 🎮 Pacman Game Clone

A classic Pacman game clone with a **Commodore 64 retro aesthetic**, built with modern web technologies.

![Pacman Game](https://img.shields.io/badge/Game-Pacman-yellow?style=for-the-badge)
![React](https://img.shields.io/badge/React-19-blue?style=for-the-badge&logo=react)
![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=for-the-badge&logo=go)
![TypeScript](https://img.shields.io/badge/TypeScript-5.9-3178C6?style=for-the-badge&logo=typescript)

## ✨ Features

- 🕹️ Classic Pacman gameplay with arrow key controls
- 👻 4 ghosts (red, pink, cyan, orange) with AI movement
- ⚡ Power pellet mechanic with 5-second power mode
- 🎯 Score tracking and collision detection
- 🏆 Scores persisted server-side on game over
- 🔐 User authentication (signup/login)
- 🎨 Retro C64-style visual design

## 🏗️ Project Structure

```
/
├── frontend/               # React application
│   ├── src/
│   │   ├── App.tsx              # Main app component
│   │   ├── App.css              # App styling
│   │   ├── game/
│   │   │   ├── Game.tsx         # Game screen + socket handling
│   │   │   ├── GameBoard/       # Board rendering layers
│   │   │   │   ├── index.tsx
│   │   │   │   ├── GridLayer.tsx
│   │   │   │   ├── PlayerLayer.tsx
│   │   │   │   ├── GhostLayer.tsx
│   │   │   │   └── ScoreDisplay.tsx
│   │   │   ├── Game.css         # Game styling
│   │   │   └── constants.ts     # Game constants, types, and map data
│   │   ├── components/
│   │   │   ├── AuthForm.tsx     # User authentication form
│   │   │   ├── GameButton.tsx   # Reusable game button component
│   │   │   ├── GameOverDialog.tsx # Game over dialog component
│   │   │   ├── ScoreBoard/      # High score view
│   │   │   │   ├── index.tsx
│   │   │   │   └── ScoreTable.tsx
│   │   │   └── TouchControls.tsx # Mobile touch controls
│   │   └── main.tsx             # Entry point
│   └── package.json
│
├── backend/                # Go server
│   ├── main.go                  # HTTP server with API endpoints
│   ├── go.mod
│   └── go.sum
│
├── infra/                  # Azure infrastructure
│   ├── main.bicep               # Main Bicep deployment file
│   └── modules/                 # Bicep modules
│
├── .github/                # GitHub Actions workflows
├── docker-compose.yml      # Local PostgreSQL database
├── Dockerfile              # Production container image
├── Makefile                # Build automation
├── start_local.sh          # Local development script
├── deploy.sh               # Deployment script
├── AGENTS.md               # AI agent instructions
└── README.md               # This file
```

## 🚀 Quick Start

### Prerequisites

- Node.js 18+
- Go 1.22+
- Docker (for local database)

### Local Development

1. **Start the database:**
   ```bash
   docker-compose up -d
   ```

2. **Create environment file (first time only):**
   ```bash
   cp .env.example .env
   # Edit .env with your settings if needed
   ```

3. **Start the backend:**
   ```bash
   make run-backend-dev
   ```

4. **Start the frontend (in a new terminal):**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

5. Open http://localhost:5173 in your browser

### Using the Start Script

```bash
./start_local.sh
```

## 🛠️ Development Commands

### Frontend

```bash
cd frontend
npm install        # Install dependencies
npm run dev        # Start dev server (http://localhost:5173)
npm run build      # Build for production
npm run lint       # Run ESLint
npm run preview    # Preview production build
```

### Backend

```bash
cd backend
go run main.go     # Run server (http://localhost:6060)
```

### Make Commands

```bash
make build-frontend    # Build frontend for production
make run-backend       # Run the Go backend server
make dev-frontend      # Run Vite dev server
make all               # Build frontend and run backend
```

## 🎮 Game Mechanics

### Controls
- **Arrow Keys**: Move Pacman (Up, Down, Left, Right)

### Game Elements
| Symbol | Element        | Description                           |
|--------|---------------|---------------------------------------|
| Wall   | Blue blocks   | Impassable barriers                   |
| Dot    | Small dots    | Collect for points                    |
| Power  | Large pellets | Enables ghost eating for 5 seconds    |
| Ghost  | Colored ghosts| Avoid or eat when powered up          |

### Map Grid
- **21 rows × 19 columns**
- Cell values in `constants.ts`:
  - `0` = Empty
  - `1` = Wall
  - `2` = Dot
  - `3` = Power Pellet
  - `9` = Ghost house door

## 🔌 API Endpoints

| Method | Endpoint      | Description                    |
|--------|--------------|--------------------------------|
| GET    | `/api/score` | Get high score                 |
| POST   | `/api/signup`| Create new user account        |
| POST   | `/api/login` | Authenticate existing user     |

## 🗄️ Database

The application uses PostgreSQL for user authentication.

**Local Configuration (Docker):**
- Host: `localhost`
- Port: `5434`
- User: `postgres`
- Password: `password`
- Database: `pacmangame`

## 🏛️ Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌────────────────┐
│  React Frontend │────▶│   Go Backend     │────▶│  PostgreSQL    │
│  (Vite + TS)    │     │   (Port 6060)    │     │  (Port 5434)   │
└─────────────────┘     └──────────────────┘     └────────────────┘
```

## 🔒 Security

### Authentication
- User passwords are hashed with bcrypt
- Session tokens are issued on login/signup
- WebSocket connections require valid session tokens

### WebSocket Security
- Origin validation using `ALLOWED_ORIGINS` environment variable
- Session-based authentication prevents impersonation
- Tokens expire after 24 hours

### Environment Variables
| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | PostgreSQL host | (required) |
| `DB_PORT` | PostgreSQL port | 5432 |
| `DB_USER` | Database user | (required) |
| `DB_PASSWORD` | Database password | (required) |
| `DB_NAME` | Database name | (required) |
| `DB_SSLMODE` | SSL mode | require |
| `ALLOWED_ORIGINS` | Comma-separated allowed origins | localhost URLs |

## 📦 Deployment

### Azure Infrastructure
Bicep templates for Azure deployment are in the `infra/` directory.

**GitHub Variables for deployment:**
- `AZURE_RESOURCE_GROUP` - Resource group name
- `AZURE_LOCATION` - Azure region (default: swedencentral)
- `ALLOWED_ORIGINS` - (Optional) Override allowed origins for WebSocket CORS

### Docker
```bash
docker build -t pacman-game .
docker run -p 6060:6060 \
  -e DB_HOST=your-db-host \
  -e DB_USER=your-user \
  -e DB_PASSWORD=your-password \
  -e DB_NAME=pacmangame \
  -e ALLOWED_ORIGINS=https://your-domain.com \
  pacman-game
```

## 📄 License

This project is private and not licensed for public use.

---

*Built with ❤️ using React, TypeScript, Go, and a love for retro games*
