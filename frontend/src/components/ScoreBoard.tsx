import React, { useEffect, useState } from 'react';
import GameButton from './GameButton';
import './ScoreBoard.css';

interface ScoreEntry {
    nickname: string;
    score: number;
}

interface ScoreBoardProps {
    onBack: () => void;
}

const ScoreBoard: React.FC<ScoreBoardProps> = ({ onBack }) => {
    const [scores, setScores] = useState<ScoreEntry[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        fetch('/api/scoreboard')
            .then(res => {
                if (!res.ok) throw new Error('Failed to load scoreboard');
                return res.json();
            })
            .then((data: ScoreEntry[]) => {
                setScores(data || []);
                setLoading(false);
            })
            .catch(err => {
                console.error('Scoreboard error:', err);
                setError('Failed to load scoreboard');
                setLoading(false);
            });
    }, []);

    return (
        <div className="scoreboard-container">
            <h2 className="scoreboard-title">*** HIGH SCORES ***</h2>
            
            {loading && <div className="scoreboard-loading">LOADING...</div>}
            
            {error && <div className="scoreboard-error">{error}</div>}
            
            {!loading && !error && (
                <div className="scoreboard-table">
                    <div className="scoreboard-header">
                        <span className="rank-col">RANK</span>
                        <span className="name-col">PLAYER</span>
                        <span className="score-col">SCORE</span>
                    </div>
                    {scores.length === 0 ? (
                        <div className="scoreboard-empty">NO SCORES YET</div>
                    ) : (
                        scores.map((entry, index) => (
                            <div key={entry.nickname} className="scoreboard-row">
                                <span className="rank-col">{index + 1}.</span>
                                <span className="name-col">{entry.nickname}</span>
                                <span className="score-col">{entry.score}</span>
                            </div>
                        ))
                    )}
                </div>
            )}
            
            <GameButton 
                onClick={onBack} 
                label="Back to Game" 
                className="back-btn"
            />
        </div>
    );
};

export default ScoreBoard;
