import React from 'react';
import type { Direction } from '../game/constants';
import './TouchControls.css';

interface TouchControlsProps {
    onDirectionChange: (dir: Direction) => void;
}

const TouchControls: React.FC<TouchControlsProps> = ({ onDirectionChange }) => {
    // Prevent default touch behavior (scrolling/zooming) when interacting with controls
    const handleTouchStart = (e: React.TouchEvent, dir: Direction) => {
        e.preventDefault(); // crucial for preventing scroll on mobile
        onDirectionChange(dir);
    };

    const handleMouseDown = (e: React.MouseEvent, dir: Direction) => {
        e.preventDefault();
        onDirectionChange(dir);
    }

    return (
        <div className="touch-controls">
            <div className="d-pad">
                <button 
                    className="d-btn up" 
                    onTouchStart={(e) => handleTouchStart(e, 'UP')}
                    onMouseDown={(e) => handleMouseDown(e, 'UP')}
                    aria-label="Up"
                >
                    ▲
                </button>
                <button 
                    className="d-btn left" 
                    onTouchStart={(e) => handleTouchStart(e, 'LEFT')}
                    onMouseDown={(e) => handleMouseDown(e, 'LEFT')}
                    aria-label="Left"
                >
                    ◀
                </button>
                <div className="d-center"></div>
                <button 
                    className="d-btn right" 
                    onTouchStart={(e) => handleTouchStart(e, 'RIGHT')}
                    onMouseDown={(e) => handleMouseDown(e, 'RIGHT')}
                    aria-label="Right"
                >
                    ▶
                </button>
                <button 
                    className="d-btn down" 
                    onTouchStart={(e) => handleTouchStart(e, 'DOWN')}
                    onMouseDown={(e) => handleMouseDown(e, 'DOWN')}
                    aria-label="Down"
                >
                    ▼
                </button>
            </div>
        </div>
    );
};

export default TouchControls;
