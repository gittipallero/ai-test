import React from 'react';
import './GameButton.css';

interface GameButtonProps {
    onClick: () => void;
    label: string;
    className?: string;
}

const GameButton: React.FC<GameButtonProps> = ({ onClick, label, className = '' }) => {
    return (
        <button 
            className={`game-button ${className}`} 
            onClick={onClick}
        >
            {label}
        </button>
    );
};

export default GameButton;
