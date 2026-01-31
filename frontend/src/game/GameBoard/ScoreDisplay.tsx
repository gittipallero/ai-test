import React from 'react';
import type { GameMode } from '../constants';

interface ScoreDisplayProps {
    score: number;
    gameMode: GameMode;
}

const ScoreDisplay: React.FC<ScoreDisplayProps> = ({ score, gameMode }) => (
    <div className="score-display">
        {gameMode === 'pair' ? 'PAIR SCORE: ' : 'SCORE: '}
        {score}
    </div>
);

export default ScoreDisplay;
