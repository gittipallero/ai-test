import React from 'react';
import GameButton from './GameButton';
import './GameOverDialog.css';

interface GameOverDialogProps {
    onRestart: () => void;
    onLogout: () => void;
}

const GameOverDialog: React.FC<GameOverDialogProps> = ({ onRestart, onLogout }) => {
    return (
        <div className="game-over">
            <div className="game-over-text">GAME OVER</div>
            <GameButton 
                onClick={onRestart} 
                label="Start New Game" 
                className="restart-btn"
            />
            <GameButton 
                onClick={onLogout} 
                label="Logout" 
                className="logout-btn"
            />
        </div>
    );
};

export default GameOverDialog;
