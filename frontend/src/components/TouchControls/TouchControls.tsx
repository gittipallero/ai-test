import React from 'react';
import type { Direction } from '../../game/constants';
import './TouchControls.css';
import ArrowButton from '../ArrowButton/ArrowButton';

interface TouchControlsProps {
    onDirectionChange: (dir: Direction) => void;
}

const TouchControls: React.FC<TouchControlsProps> = ({ onDirectionChange }) => {
    return (
        <div className="touch-controls">
            <div className="d-pad">
                <ArrowButton 
                    className="up" 
                    onTrigger={() => onDirectionChange('UP')}
                    label="Up"
                >
                    ▲
                </ArrowButton>
                <ArrowButton 
                    className="left" 
                    onTrigger={() => onDirectionChange('LEFT')}
                    label="Left"
                >
                    ◀
                </ArrowButton>
                <div className="d-center"></div>
                <ArrowButton 
                    className="right" 
                    onTrigger={() => onDirectionChange('RIGHT')}
                    label="Right"
                >
                    ▶
                </ArrowButton>
                <ArrowButton 
                    className="down" 
                    onTrigger={() => onDirectionChange('DOWN')}
                    label="Down"
                >
                    ▼
                </ArrowButton>
            </div>
        </div>
    );
};

export default TouchControls;
