import React, { useState, useEffect } from 'react';
import { INITIAL_MAP, ROWS, COLS, BLOCK_SIZE, INITIAL_GHOSTS, INITIAL_PACMAN } from './constants';
import type { Direction, Position, GhostEntity } from './constants';
import GameButton from '../components/GameButton';
import './Game.css';

const Game: React.FC = () => {
    // Deep copy map to handle state (eating dots)
    const [grid, setGrid] = useState<number[][]>(INITIAL_MAP.map(row => [...row]));
    const [pacman, setPacman] = useState<Position>(INITIAL_PACMAN);
    const [ghosts, setGhosts] = useState<GhostEntity[]>(INITIAL_GHOSTS);
    const [direction, setDirection] = useState<Direction>(null);
    const [nextDirection, setNextDirection] = useState<Direction>(null);
    const [score, setScore] = useState(0);
    const [gameOver, setGameOver] = useState(false);
    const [powerModeTime, setPowerModeTime] = useState(0);

    // Initial check for counting dots could be done here to know when win

    // Handle Input
    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (gameOver) return;
            switch(e.key) {
                case 'ArrowUp': setNextDirection('UP'); break;
                case 'ArrowDown': setNextDirection('DOWN'); break;
                case 'ArrowLeft': setNextDirection('LEFT'); break;
                case 'ArrowRight': setNextDirection('RIGHT'); break;
            }
        };
        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [gameOver]);

    // Game Loop
    useEffect(() => {
        if (gameOver) return;
        const interval = setInterval(() => {
            movePacman();
            moveGhosts();
            checkCollisions();
            if (powerModeTime > 0) {
                setPowerModeTime(t => t - 150);
            }
        }, 150); // Tick rate
        return () => clearInterval(interval);
    }, [pacman, direction, nextDirection, gameOver, ghosts, powerModeTime, grid]);

    const movePacman = () => {
        let currentDir = direction;
        
        // Try to change direction if possible
        if (nextDirection && canMove(pacman, nextDirection)) {
            currentDir = nextDirection;
            setDirection(nextDirection);
        }

        if (currentDir && canMove(pacman, currentDir)) {
            const newPos = getNextPos(pacman, currentDir);
            
            // Check for portal
            if (newPos.x < 0) newPos.x = COLS - 1;
            if (newPos.x >= COLS) newPos.x = 0;

            // Eat Dot
            if (grid[newPos.y][newPos.x] === 2) {
                const newGrid = [...grid];
                newGrid[newPos.y] = [...newGrid[newPos.y]];
                newGrid[newPos.y][newPos.x] = 0;
                setGrid(newGrid);
                setScore(s => s + 10);
            }
            // Eat Power Pellet
            if (grid[newPos.y][newPos.x] === 3) {
                const newGrid = [...grid];
                newGrid[newPos.y] = [...newGrid[newPos.y]];
                newGrid[newPos.y][newPos.x] = 0;
                setGrid(newGrid);
                setScore(s => s + 50);
                setPowerModeTime(5000); // 5 seconds of power mode
            }

            setPacman(newPos);
        }
    };

    const getNextPos = (pos: Position, dir: Direction): Position => {
        const newPos = { ...pos };
        if (dir === 'UP') newPos.y -= 1;
        if (dir === 'DOWN') newPos.y += 1;
        if (dir === 'LEFT') newPos.x -= 1;
        if (dir === 'RIGHT') newPos.x += 1;
        return newPos;
    };

    const canMove = (pos: Position, dir: Direction): boolean => {
        const next = getNextPos(pos, dir);
        // Strict Y Check first
        if (next.y < 0 || next.y >= ROWS) return false;
        // Tunnel handling for boundary checks
        if (next.x < 0 || next.x >= COLS) return true; 
        return grid[next.y][next.x] !== 1;
    };

    const moveGhosts = () => {
        setGhosts(prevGhosts => prevGhosts.map(ghost => {
            let possibleDirs: Direction[] = ['UP', 'DOWN', 'LEFT', 'RIGHT'];
            let validDirs = possibleDirs.filter(d => canMove(ghost.pos, d));
            
            // Simple logic: continue current direction if possible, otherwise mostly random
            // Or just random for now to avoid getting stuck easily
            
            // Filter out reverse direction if we have other options to prevent jitter
            const reverseDir = getReverseDir(ghost.dir);
            if (validDirs.length > 1 && ghost.dir) {
                const nonReverse = validDirs.filter(d => d !== reverseDir);
                if (nonReverse.length > 0) validDirs = nonReverse;
            }

            let nextDir = ghost.dir;
            if (!ghost.dir || !canMove(ghost.pos, ghost.dir) || Math.random() < 0.2) {
                 if (validDirs.length > 0) {
                     nextDir = validDirs[Math.floor(Math.random() * validDirs.length)];
                 }
            }

            if (nextDir && canMove(ghost.pos, nextDir)) {
                 let newPos = getNextPos(ghost.pos, nextDir);
                 // Check for portal
                 if (newPos.x < 0) newPos.x = COLS - 1;
                 if (newPos.x >= COLS) newPos.x = 0;

                 return { ...ghost, pos: newPos, dir: nextDir };
            }
             return ghost;
        }));
    };

    const getReverseDir = (dir: Direction): Direction => {
        if (dir === 'UP') return 'DOWN';
        if (dir === 'DOWN') return 'UP';
        if (dir === 'LEFT') return 'RIGHT';
        if (dir === 'RIGHT') return 'LEFT';
        return null;
    };

    const checkCollisions = () => {
        // Need to check current positions
        // Note: pacman state here is from closure, might be stale in setInterval if not dependencies updated
        // We added dependencies to useEffect so it resets interval on state change, which works for this simple loop
        // But for better performance we should use refs or functional updates. 
        // With functional updates, checking collision is tricky.
        // Let's do a simple check here based on current render state which is what the effect sees.
        
        // Actually, since we update state, the effect re-runs.
        // We can check collision against the new positions?
        // Let's rely on the effect dependency update for now. 
        // We need to check if ANY ghost is at pacman position.
    
       // We can iterate ghosts state (but we don't have the *new* ghost state here immediately after setGhosts)
       // So we should probably check collision in a separate effect or use a ref for updated positions.
       // However, for simplicity, checking in the next tick or just using the values available is fine.
       // Let's check against the *current* state (which was used for previous frame).
       // Or better: check collision in render? No, that's side effect.
       
       // Let's simplify: check collision inside the interval using refs would be best, but I'll stick to state.
       // The interval clears and restarts on every state change because of dependency array. 
       // This is inefficient but fine for this simple game.
       
       // Using the *ghosts* state from closure (which is current render's state).
       // If a ghost occupies the same tile as pacman.
       const hitGhost = ghosts.find(g => g.pos.x === pacman.x && g.pos.y === pacman.y);
       if (hitGhost) {
           if (powerModeTime > 0) {
               // Eat ghost
               setScore(s => s + 200);
               // Send ghost back home
               setGhosts(prev => prev.map(g => g.id === hitGhost.id ? { ...g, pos: { x: 9, y: 8 } } : g));
           } else {
               setGameOver(true);
           }
       }
    };

    const resetGame = () => {
        setGrid(INITIAL_MAP.map(row => [...row]));
        setPacman(INITIAL_PACMAN);
        setGhosts(INITIAL_GHOSTS);
        setDirection(null);
        setNextDirection(null);
        setScore(0);
        setGameOver(false);
        setPowerModeTime(0);
    };

    return (
        <div className="game-board" style={{ width: COLS * BLOCK_SIZE, height: ROWS * BLOCK_SIZE }}>
            {grid.map((row, y) => (
                row.map((cell, x) => (
                    <div key={`${x}-${y}`} className={`cell cell-${cell}`} style={{
                        left: x * BLOCK_SIZE,
                        top: y * BLOCK_SIZE,
                        width: BLOCK_SIZE,
                        height: BLOCK_SIZE
                    }}></div>
                ))
            ))}
            <div className={`pacman pacman-${direction}`} style={{
                left: pacman.x * BLOCK_SIZE,
                top: pacman.y * BLOCK_SIZE,
                width: BLOCK_SIZE,
                height: BLOCK_SIZE
            }}></div>
            {ghosts.map(ghost => (
                <div key={ghost.id} className="ghost" style={{
                    left: ghost.pos.x * BLOCK_SIZE,
                    top: ghost.pos.y * BLOCK_SIZE,
                    width: BLOCK_SIZE,
                    height: BLOCK_SIZE,
                    backgroundColor: powerModeTime > 0 ? 'blue' : ghost.color
                }}></div>
            ))}
            <div className="score-display">SCORE: {score}</div>
            {gameOver && (
                <div className="game-over">
                    <div className="game-over-text">GAME OVER</div>
                    <GameButton 
                        onClick={resetGame} 
                        label="Start New Game" 
                        className="restart-btn"
                    />
                </div>
            )}
        </div>
    );
};

export default Game;
