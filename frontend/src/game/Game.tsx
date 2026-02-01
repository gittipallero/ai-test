import React, { useState, useEffect, useRef, useCallback } from 'react';
import { COLS, BLOCK_SIZE } from './constants';
import type { Direction, GameMode, GameState, LobbyStats } from './constants';
import GameBoard from './GameBoard';
import GameOverDialog from '../components/GameOverDialog';
import TouchControls from '../components/TouchControls';

import './Game.css';

interface GameProps {
    onLogout: () => void;
    onShowScoreboard: () => void;
    onOnlineCountChange: (count: number) => void;
    username: string;
    authToken: string;
}

const Game: React.FC<GameProps> = ({ onLogout, onShowScoreboard, onOnlineCountChange, username, authToken }) => {
    const [gameState, setGameState] = useState<GameState | null>(null);
    const [waiting, setWaiting] = useState(false);
    const [lobbyStats, setLobbyStats] = useState<LobbyStats>({ online_count: 0 });
    const [gameMode, setGameMode] = useState<GameMode>(null);
    const [scale, setScale] = useState(1);
    
    // Derived state for local player
    const [localDirection, setLocalDirection] = useState<Direction>(null);

    const ws = useRef<WebSocket | null>(null);

    // Scaling logic
    useEffect(() => {
        const handleResize = () => {
            const boardWidth = COLS * BLOCK_SIZE; // 380
            // Add some padding/margin consideration (e.g. 40px)
            const availableWidth = window.innerWidth - 20; 
            const newScale = Math.min(availableWidth / boardWidth, 1);
            setScale(newScale);
        };

        handleResize(); // Initial call
        window.addEventListener('resize', handleResize);
        return () => window.removeEventListener('resize', handleResize);
    }, []);

    useEffect(() => {
        // Connect to WebSocket with session token for authentication
        const wsProtocol = window.location.protocol === 'https:' ? 'wss' : 'ws';
        const wsHost = window.location.host;
        const wsUrl = `${wsProtocol}://${wsHost}/api/ws?token=${encodeURIComponent(authToken)}`;
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
    }, [authToken, onOnlineCountChange]);

    const handleDirectionInput = useCallback((dir: Direction) => {
        const currentSocket = ws.current;
        if (gameState?.gameOver || !currentSocket || currentSocket.readyState !== WebSocket.OPEN) {
            return;
        }
        if (dir) {
            setLocalDirection(dir);
            currentSocket.send(JSON.stringify({ type: 'input', direction: dir }));
        }
    }, [gameState?.gameOver]);

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            let dir: Direction = null;
            switch(e.key) {
                case 'ArrowUp': dir = 'UP'; break;
                case 'ArrowDown': dir = 'DOWN'; break;
                case 'ArrowLeft': dir = 'LEFT'; break;
                case 'ArrowRight': dir = 'RIGHT'; break;
            }
            if (dir) {
                handleDirectionInput(dir);
            }
        };
        window.addEventListener('keydown', handleKeyDown);
        return () => window.removeEventListener('keydown', handleKeyDown);
    }, [handleDirectionInput]);

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
            <div className="game-wrapper">
                <div className="game-board-centered">
                    <div className="waiting-screen">
                        <h2>Waiting for opponent...</h2>
                        <p>Online users: {lobbyStats.online_count}</p>
                    </div>
                </div>
                <TouchControls onDirectionChange={handleDirectionInput} />
            </div>
        );
    }

    if (!gameState) {
        return (
            <div className="game-wrapper">
                <div className="game-loading">Loading...</div>
                <TouchControls onDirectionChange={handleDirectionInput} />
            </div>
        );
    }

    const { grid, players, ghosts, score, gameOver, powerModeTime } = gameState;
    
    return (
        <div className="game-wrapper">
             <div className="game-header">
                {/* Online count moved to App header */}
            </div>
            <GameBoard
                grid={grid}
                players={players}
                ghosts={ghosts}
                score={score}
                gameMode={gameMode}
                powerModeTime={powerModeTime}
                localDirection={localDirection}
                username={username}
                scale={scale}
            >
                {gameOver && (
                    <GameOverDialog 
                        onRestart={handleRestart} 
                        onLogout={onLogout}
                        onShowScoreboard={onShowScoreboard}
                        onStartPairGame={handleStartPairGame}
                        showPairButton={lobbyStats.online_count > 1}
                        score={score}
                    />
                )}
            </GameBoard>

            <TouchControls onDirectionChange={handleDirectionInput} />
        </div>
    );
};

export default Game;
