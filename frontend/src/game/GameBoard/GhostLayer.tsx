import React from 'react';
import { BLOCK_SIZE } from '../constants';
import type { GhostEntity } from '../constants';

interface GhostLayerProps {
    ghosts: GhostEntity[];
    powerModeTime: number;
}

const GhostLayer: React.FC<GhostLayerProps> = ({ ghosts, powerModeTime }) => (
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

export default GhostLayer;
