import React from 'react';
import GameButton from '../GameButton/GameButton';
import './GameOverDialog.css';

interface GameOverDialogProps {
    onRestart: () => void;
    onLogout: () => void;
    onShowScoreboard: () => void;
    onStartPairGame: () => void;
    showPairButton: boolean;
    score: number;
}

const GameOverDialog: React.FC<GameOverDialogProps> = ({ 
    onRestart, 
    onLogout, 
    onShowScoreboard,
    onStartPairGame,
    showPairButton,
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
            {showPairButton && (
                <GameButton 
                    onClick={onStartPairGame} 
                    label="Start Pair Game" 
                    className="pair-setup-btn"
                />
            )}
            <GameButton 
                onClick={onLogout} 
                label="Logout" 
                className="logout-btn"
            />
        </div>
    );
};

export default GameOverDialog;

