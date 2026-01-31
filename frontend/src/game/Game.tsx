import React, { useState, useEffect, useRef } from 'react';
import { INITIAL_MAP, ROWS, COLS, BLOCK_SIZE, INITIAL_PACMAN, INITIAL_GHOSTS } from './constants';
import type { Direction, Position, GhostEntity } from './constants';
import GameOverDialog from '../components/GameOverDialog';
import './Game.css';

interface GameProps {
    onLogout: () => void;
    onShowScoreboard: () => void;
    username: string;
}

interface GameState {
    grid: number[][];
    pacman: Position;
    ghosts: GhostEntity[];
    score: number;
    gameOver: boolean;
    powerModeTime: number;
    direction: Direction;
    nextDirection: Direction;
}

const Game: React.FC<GameProps> = ({ onLogout, onShowScoreboard, username }) => {
    const [gameState, setGameState] = useState<GameState>({
        grid: INITIAL_MAP,
        pacman: INITIAL_PACMAN,
        ghosts: INITIAL_GHOSTS,
        score: 0,
        gameOver: false,
        powerModeTime: 0,
        direction: null,
        nextDirection: null,
    });
    const [gameId, setGameId] = useState(0);
    const ws = useRef<WebSocket | null>(null);

    useEffect(() => {
        // Connect to WebSocket
        const socket = new WebSocket(`ws://localhost:6060/api/ws?nickname=${encodeURIComponent(username)}`);
        
        socket.onopen = () => {
            console.log('Connected to game server');
        };

        socket.onmessage = (event) => {
            try {
                const newState: GameState = JSON.parse(event.data);
                setGameState(newState);
            } catch (e) {
                console.error('Error parsing game state', e);
            }
        };

        socket.onclose = () => {
            console.log('Disconnected from game server');
        };

        ws.current = socket;

        return () => {
            socket.close();
        };
    }, [username, gameId]);

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (gameState.gameOver || !ws.current) return;
            let dir: Direction = null;
            switch(e.key) {
                case 'ArrowUp': dir = 'UP'; break;
                case 'ArrowDown': dir = 'DOWN'; break;
                case 'ArrowLeft': dir = 'LEFT'; break;
                case 'ArrowRight': dir = 'RIGHT'; break;
            }
            if (dir) {
                ws.current.send(JSON.stringify({ direction: dir }));
            }
        };
        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [gameState.gameOver]);

    const handleRestart = () => {
         setGameId(prev => prev + 1);
    };

    const { grid, pacman, ghosts, direction, score, gameOver, powerModeTime } = gameState;

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
                <GameOverDialog 
                    onRestart={handleRestart} 
                    onLogout={onLogout}
                    onShowScoreboard={onShowScoreboard}
                    score={score}
                    nickname={username}
                />
            )}
        </div>
    );
};

export default Game;
