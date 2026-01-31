import React from 'react';
import { BLOCK_SIZE } from '../constants';
import type { Direction, PlayerState } from '../constants';

interface PlayerLayerProps {
    players: Record<string, PlayerState>;
    localDirection: Direction;
    username: string;
}

const PlayerLayer: React.FC<PlayerLayerProps> = ({ players, localDirection, username }) => (
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

export default PlayerLayer;
