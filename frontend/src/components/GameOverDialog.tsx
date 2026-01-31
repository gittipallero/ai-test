import React from 'react';
import GameButton from './GameButton';
import './GameOverDialog.css';

interface GameOverDialogProps {
    onRestart: () => void;
    onLogout: () => void;
    onShowScoreboard: () => void;
    score: number;
}

const GameOverDialog: React.FC<GameOverDialogProps> = ({ 
    onRestart, 
    onLogout, 
    onShowScoreboard,
    score
}) => {
    return (
        <div className="game-over">
            <div className="game-over-text">GAME OVER</div>
            <div className="final-score">SCORE: {score}</div>
            <GameButton 
                onClick={onRestart} 
                label="Start New Game" 
                className="restart-btn"
            />
            <GameButton 
                onClick={onShowScoreboard} 
                label="Score Board" 
                className="scoreboard-btn"
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

