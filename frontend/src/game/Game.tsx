import React, { useState, useEffect, useRef } from 'react';
import { ROWS, COLS, BLOCK_SIZE } from './constants';
import type { Direction, Position, GhostEntity } from './constants';
import GameOverDialog from '../components/GameOverDialog';

import './Game.css';

interface GameProps {
    onLogout: () => void;
    onShowScoreboard: () => void;
    onOnlineCountChange: (count: number) => void;
    username: string;
}

interface PlayerState {
    nickname: string;
    pos: Position;
    dir: Direction;
    alive: boolean;
}

interface GameState {
    grid: number[][];
    players: Record<string, PlayerState>;
    ghosts: GhostEntity[];
    score: number;
    gameOver: boolean;
    powerModeTime: number;
}

interface LobbyStats {
    online_count: number;
}

const Game: React.FC<GameProps> = ({ onLogout, onShowScoreboard, onOnlineCountChange, username }) => {
    const [gameState, setGameState] = useState<GameState | null>(null);
    const [waiting, setWaiting] = useState(false);
    const [lobbyStats, setLobbyStats] = useState<LobbyStats>({ online_count: 0 });
    const [gameMode, setGameMode] = useState<'single' | 'pair' | null>(null);
    
    // Derived state for local player
    const [localDirection, setLocalDirection] = useState<Direction>(null);

    const ws = useRef<WebSocket | null>(null);

    useEffect(() => {
        // Connect to WebSocket
        const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
        const wsHost = window.location.host;
        const wsUrl = `${wsProtocol}://${wsHost}/api/ws?nickname=${encodeURIComponent(username)}`;
        const socket = new WebSocket(wsUrl);
        
        socket.onopen = () => {
            console.log('Connected to game server');
            socket.send(JSON.stringify({ type: 'start_single' }));
            setGameMode('single');
        };

        socket.onmessage = (event) => {
            try {
                const msg = JSON.parse(event.data);
                
                if (msg.type === 'lobby_stats') {
                    setLobbyStats({ online_count: msg.online_count });
                    onOnlineCountChange(msg.online_count);
                } else if (msg.type === 'waiting') {
                    setWaiting(true);
                    setGameMode(null);
                    setGameState(null);
                } else if (msg.type === 'game_start') {
                    setWaiting(false);
                    setGameMode(msg.mode);
                } else if (msg.grid) {
                    setGameState(msg);
                }
            } catch (e) {
                console.error('Error parsing game state', e);
            }
        };

        socket.onclose = () => {
            console.log('Disconnected from game server');
            if (ws.current === socket) {
                ws.current = null;
            }
        };

        ws.current = socket;

        return () => {
            if (ws.current === socket) {
                ws.current = null;
            }
            socket.close();
        };
    }, [username, onOnlineCountChange]);

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            const currentSocket = ws.current;
            if (gameState?.gameOver || !currentSocket || currentSocket.readyState !== WebSocket.OPEN) {
                return;
            }
            let dir: Direction = null;
            switch(e.key) {
                case 'ArrowUp': dir = 'UP'; break;
                case 'ArrowDown': dir = 'DOWN'; break;
                case 'ArrowLeft': dir = 'LEFT'; break;
                case 'ArrowRight': dir = 'RIGHT'; break;
            }
            if (dir) {
                setLocalDirection(dir);
                currentSocket.send(JSON.stringify({ type: 'input', direction: dir }));
            }
        };
        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [gameState?.gameOver]);

    const handleStartPairGame = () => {
        if (ws.current) {
            ws.current.send(JSON.stringify({ type: 'join_pair' }));
        }
    };

    const handleRestart = () => {
         if (ws.current) {
             if (gameMode === 'pair') {
                ws.current.send(JSON.stringify({ type: 'join_pair' }));
             } else {
                ws.current.send(JSON.stringify({ type: 'start_single' }));
             }
         }
    };
    
    if (waiting) {
        return (
            <div className="game-board-centered">
                <div className="waiting-screen">
                    <h2>Waiting for opponent...</h2>
                    <p>Online users: {lobbyStats.online_count}</p>
                </div>
            </div>
        );
    }

    if (!gameState) {
        return <div className="game-loading">Loading...</div>;
    }

    const { grid, players, ghosts, score, gameOver, powerModeTime } = gameState;
    
    return (
        <div className="game-wrapper">
             <div className="game-header">
                {/* Online count moved to App header */}
            </div>

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
                
                {Object.values(players).map((p) => (
                    p.alive && (
                        <div key={p.nickname} className={`pacman pacman-${p.dir || localDirection || 'RIGHT'}`} style={{
                            left: p.pos.x * BLOCK_SIZE,
                            top: p.pos.y * BLOCK_SIZE,
                            width: BLOCK_SIZE,
                            height: BLOCK_SIZE,
                            filter: p.nickname !== username ? 'hue-rotate(180deg)' : 'none'
                        }}>
                             <div className="player-name" style={{
                                 position: 'absolute',
                                 top: '-15px',
                                 left: '50%',
                                 transform: 'translateX(-50%)',
                                 fontSize: '10px',
                                 color: 'white',
                                 whiteSpace: 'nowrap'
                             }}>{p.nickname}</div>
                        </div>
                    )
                ))}

                {ghosts.map(ghost => (
                    <div key={ghost.id} className="ghost" style={{
                        left: ghost.pos.x * BLOCK_SIZE,
                        top: ghost.pos.y * BLOCK_SIZE,
                        width: BLOCK_SIZE,
                        height: BLOCK_SIZE,
                        backgroundColor: powerModeTime > 0 ? 'blue' : ghost.color
                    }}></div>
                ))}
                
                <div className="score-display">
                    {gameMode === 'pair' ? 'PAIR SCORE: ' : 'SCORE: '} 
                    {score}
                </div>
                
                {gameOver && (
                    <GameOverDialog 
                        onRestart={handleRestart} 
                        onLogout={onLogout}
                        onShowScoreboard={onShowScoreboard}
                        onStartPairGame={handleStartPairGame}
                        showPairButton={lobbyStats.online_count > 1}
                        shouldSubmitScore={gameMode === 'single'}
                        score={score}
                        nickname={username}
                    />
                )}
            </div>
        </div>
    );
};

export default Game;
