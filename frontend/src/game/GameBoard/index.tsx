import React from 'react';
import { BLOCK_SIZE, COLS, ROWS } from '../constants';
import type { Direction, GameMode, GhostEntity, PlayerState } from '../constants';
import GridLayer from './GridLayer';
import PlayerLayer from './PlayerLayer';
import GhostLayer from './GhostLayer';
import ScoreDisplay from './ScoreDisplay';

interface GameBoardProps {
    grid: number[][];
    players: Record<string, PlayerState>;
    ghosts: GhostEntity[];
    score: number;
    gameMode: GameMode;
    powerModeTime: number;
    localDirection: Direction;
    username: string;
    scale: number;
    children?: React.ReactNode;
}

const GameBoard: React.FC<GameBoardProps> = ({
    grid,
    players,
    ghosts,
    score,
    gameMode,
    powerModeTime,
    localDirection,
    username,
    scale,
    children
}) => {
    const boardWidth = COLS * BLOCK_SIZE;
    const boardHeight = ROWS * BLOCK_SIZE;

    return (
        <div className="game-board-container" style={{
            width: boardWidth,
            height: boardHeight,
            transform: `scale(${scale})`,
            transformOrigin: 'top center',
            marginBottom: (boardHeight * scale) - boardHeight
        }}>
            <div className="game-board" style={{ width: boardWidth, height: boardHeight }}>
                <GridLayer grid={grid} />
                <PlayerLayer players={players} localDirection={localDirection} username={username} />
                <GhostLayer ghosts={ghosts} powerModeTime={powerModeTime} />
                <ScoreDisplay score={score} gameMode={gameMode} />
                {children}
            </div>
        </div>
    );
};

export default GameBoard;
