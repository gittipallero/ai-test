import React, { useEffect, useState } from 'react';
import GameButton from '../GameButton/GameButton';
import ScoreTable, { type ScoreTableRow } from './ScoreTable';
import './ScoreBoard.css';

interface ScoreEntry {
    nickname: string;
    score: number;
}

interface PairScoreEntry {
    player1: string;
    player2: string;
    score: number;
}

interface ScoreBoardProps {
    onBack: () => void;
    initialGhostCount?: number;
}

const ScoreBoard: React.FC<ScoreBoardProps> = ({ onBack, initialGhostCount = 4 }) => {
    const [singleScores, setSingleScores] = useState<ScoreEntry[]>([]);
    const [pairScores, setPairScores] = useState<PairScoreEntry[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [viewGhostCount, setViewGhostCount] = useState(initialGhostCount);

    const handleGhostCountChange = (num: number) => {
        setLoading(true);
        setError(null);
        setViewGhostCount(num);
    };

    useEffect(() => {
        Promise.all([
            fetch(`/api/scoreboard?ghosts=${viewGhostCount}`).then(res => res.json()),
            fetch('/api/scoreboard/pair').then(res => res.json())
        ])
            .then(([singleData, pairData]) => {
                setSingleScores(singleData || []);
                setPairScores(pairData || []);
                setError(null);
                setLoading(false);
            })
            .catch(err => {
                console.error('Scoreboard error:', err);
                setError('Failed to load scoreboards');
                setLoading(false);
            });
    }, [viewGhostCount]);

    const singleRows: ScoreTableRow[] = singleScores.map((entry, index) => ({
        key: `${entry.nickname}-${index}`,
        rank: index + 1,
        name: entry.nickname,
        score: entry.score
    }));

    const pairRows: ScoreTableRow[] = pairScores.map((entry, index) => ({
        key: `${entry.player1}-${entry.player2}-${index}`,
        rank: index + 1,
        name: `${entry.player1} & ${entry.player2}`,
        score: entry.score,
        nameClassName: 'text-small'
    }));

    return (
        <div className="scoreboard-container">
            <h2 className="scoreboard-title">*** HIGH SCORES ***</h2>
            
            <div className="ghost-selector">
                <span className="ghost-selector-label">Ghosts:</span>
                {[1, 2, 3, 4, 5, 6, 7, 8, 9, 10].map(num => (
                    <button 
                        key={num} 
                        className={`ghost-btn ${viewGhostCount === num ? 'active' : ''}`}
                        onClick={() => handleGhostCountChange(num)}
                    >
                        {num}
                    </button>
                ))}
            </div>
            
            {loading && <div className="scoreboard-loading">LOADING...</div>}
            {error && <div className="scoreboard-error">{error}</div>}
            
            {!loading && !error && (
                <div className="scoreboards-wrapper">
                    <ScoreTable title="SINGLE PLAYER" nameHeader="PLAYER" rows={singleRows} />
                    <ScoreTable title="PAIR MODE" nameHeader="PLAYERS" rows={pairRows} />
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
