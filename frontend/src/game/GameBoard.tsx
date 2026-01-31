import React from 'react';
import { BLOCK_SIZE, COLS, ROWS } from './constants';
import type { Direction, GameMode, GhostEntity, PlayerState } from './constants';

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

const GridLayer: React.FC<{ grid: number[][] }> = ({ grid }) => (
    <>
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
    </>
);

const PlayerLayer: React.FC<{
    players: Record<string, PlayerState>;
    localDirection: Direction;
    username: string;
}> = ({ players, localDirection, username }) => (
    <>
        {Object.values(players).map((player) => (
            player.alive ? (
                <div key={player.nickname} className={`pacman pacman-${player.dir || localDirection || 'RIGHT'}`} style={{
                    left: player.pos.x * BLOCK_SIZE,
                    top: player.pos.y * BLOCK_SIZE,
                    width: BLOCK_SIZE,
                    height: BLOCK_SIZE,
                    filter: player.nickname !== username ? 'hue-rotate(180deg)' : 'none'
                }}>
                    <div className="player-name">{player.nickname}</div>
                </div>
            ) : null
        ))}
    </>
);

const GhostLayer: React.FC<{
    ghosts: GhostEntity[];
    powerModeTime: number;
}> = ({ ghosts, powerModeTime }) => (
    <>
        {ghosts.map((ghost) => (
            <div key={ghost.id} className="ghost" style={{
                left: ghost.pos.x * BLOCK_SIZE,
                top: ghost.pos.y * BLOCK_SIZE,
                width: BLOCK_SIZE,
                height: BLOCK_SIZE,
                backgroundColor: powerModeTime > 0 ? 'blue' : ghost.color
            }}></div>
        ))}
    </>
);

const ScoreDisplay: React.FC<{ score: number; gameMode: GameMode }> = ({ score, gameMode }) => (
    <div className="score-display">
        {gameMode === 'pair' ? 'PAIR SCORE: ' : 'SCORE: '}
        {score}
    </div>
);

export default GameBoard;
