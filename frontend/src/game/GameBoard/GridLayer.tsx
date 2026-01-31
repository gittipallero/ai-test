import React from 'react';
import { BLOCK_SIZE } from '../constants';

interface GridLayerProps {
    grid: number[][];
}

const GridLayer: React.FC<GridLayerProps> = ({ grid }) => (
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

export default GridLayer;
