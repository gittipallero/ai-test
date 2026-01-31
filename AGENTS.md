# AGENTS.md

Instructions and context for AI agents working on this codebase.

> [!IMPORTANT]
> **Also read [README.md](README.md)** for basic information about the game, features, and project structure.
>
> **Whenever the project structure or game logic changes, you MUST update README.md** to reflect those changes.

## Project Overview

This is a **Pacman game clone** with a Commodore 64 retro aesthetic. The project consists of:

- **Frontend**: React 19 + TypeScript + Vite
- **Backend**: Go HTTP server

## Development Commands

### Using Make

```bash
make build-frontend    # Build frontend for production
make run-backend       # Run the Go backend server (port 6060)
make dev-frontend      # Run Vite dev server
make all               # Build frontend and run backend
```

### Frontend (npm)

```bash
cd frontend
npm install            # Install dependencies
npm run dev            # Start dev server
npm run build          # Build for production (tsc + vite build)
npm run lint           # Run ESLint
npm run preview        # Preview production build
```

### Backend (Go)

```bash
cd backend
go run main.go         # Run server on port 6060
```

## Technical Details

### Frontend

- **React 19** with functional components and hooks
- **TypeScript** for type safety
- **Vite 7** for bundling and dev server
- **ESLint** for linting

### Backend

- **Go 1.22**
- Serves static files from `frontend/dist`
- API endpoint: `GET /api/score` returns `{ highScore: number }`
- Includes security headers (X-Frame-Options, CSP, etc.)
- Runs on port **6060**

### Game Architecture

The game logic is in `frontend/src/game/`:

- **Game.tsx**: Main game component with:
  - Grid-based movement system
  - Pacman controlled via arrow keys
  - Ghost AI with simple random movement
  - Power pellet mechanic (5-second power mode)
  - Collision detection
  - Score tracking

- **constants.ts**: Game data including:
  - `INITIAL_MAP`: 21x19 grid where:
    - `0` = Empty
    - `1` = Wall
    - `2` = Dot
    - `3` = Power Pellet
    - `9` = Ghost house door
  - `BLOCK_SIZE`: 20 pixels
  - `INITIAL_GHOSTS`: 4 ghosts (red, pink, cyan, orange)
  - TypeScript types: `Direction`, `Position`, `GhostEntity`

## Code Conventions

- Use TypeScript strict mode
- Functional React components with hooks
- Use `type` imports for TypeScript types: `import type { ... }`
- CSS files co-located with components
- Game state managed with React `useState`
- **Prefer modularity**: Break large components into smaller, reusable components. Each component should have a single responsibility. Avoid monolithic components—extract dialogs, overlays, and repeated UI patterns into their own component files.

## Testing & Quality

### Before Committing

1. Run linter: `npm run lint` (in frontend/)
2. Build frontend: `npm run build` (in frontend/)
3. Ensure no TypeScript errors

## Common Tasks

### Adding a New Game Feature

1. Modify game state in `Game.tsx`
2. Add types to `constants.ts` if needed
3. Update game loop in `useEffect` hook
4. Run lint and build to verify

### Modifying the Map

Edit `INITIAL_MAP` in `frontend/src/game/constants.ts`. Ensure the map is 21 rows × 19 columns.

### Adding API Endpoints

Add new handlers in `backend/main.go` using `mux.HandleFunc()`.

## Known Patterns

- Game uses a tick-based loop (150ms interval)
- Direction changes are queued via `nextDirection` state
- Portals (tunnels) wrap around at map edges
- Ghost collision during power mode sends ghost back to spawn
Preference: Small Functions over Long Functions
When writing code, prefer breaking down long functions into smaller, more manageable helper functions. This improves readability and maintainability.
