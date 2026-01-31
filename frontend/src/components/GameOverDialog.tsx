import React, { useEffect } from 'react';
import GameButton from './GameButton';
import './GameOverDialog.css';

interface GameOverDialogProps {
    onRestart: () => void;
    onLogout: () => void;
    onShowScoreboard: () => void;
    onStartPairGame: () => void;
    showPairButton: boolean;
    score: number;
    nickname: string;
}

const GameOverDialog: React.FC<GameOverDialogProps> = ({ 
    onRestart, 
    onLogout, 
    onShowScoreboard,
    onStartPairGame,
    showPairButton,
    score,
    nickname
}) => {
    // Submit score when game over dialog mounts
    useEffect(() => {
        if (nickname && score > 0) {
            fetch('/api/score', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ nickname, score })
            })
            .then(res => {
                if (!res.ok) console.error('Failed to submit score');
            })
            .catch(err => console.error('Score submission error:', err));
        }
    }, [nickname, score]);

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

