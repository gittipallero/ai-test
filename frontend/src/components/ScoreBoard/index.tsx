import React, { useEffect, useState } from 'react';
import GameButton from '../GameButton/GameButton';
import ScoreTable, { type ScoreTableRow } from './ScoreTable';
import type { GameMode } from '../../game/constants';
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
    activeMode?: GameMode;
}

const ScoreBoard: React.FC<ScoreBoardProps> = ({ onBack, initialGhostCount = 4, activeMode = 'single' }) => {
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
        setLoading(true);

        const fetchScores = async () => {
            try {
                if (activeMode === 'single') {
                    const response = await fetch(`/api/scoreboard?ghosts=${viewGhostCount}`);
                    const data = await response.json();
                    setSingleScores(data || []);
                } else if (activeMode === 'pair') {
                    const response = await fetch('/api/scoreboard/pair');
                    const data = await response.json();
                    setPairScores(data || []);
                }
                setError(null);
            } catch (err) {
                console.error('Scoreboard error:', err);
                setError('Failed to load scoreboards');
            } finally {
                setLoading(false);
            }
        };

        fetchScores();
    }, [viewGhostCount, activeMode]);

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
            
            {activeMode === 'single' && (
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
            )}
            
            {loading && <div className="scoreboard-loading">LOADING...</div>}
            {error && <div className="scoreboard-error">{error}</div>}
            
            {!loading && !error && (
                <div className="scoreboards-wrapper">
                    {activeMode === 'single' && (
                         <ScoreTable title="SINGLE PLAYER" nameHeader="PLAYER" rows={singleRows} />
                    )}
                    {activeMode === 'pair' && (
                        <ScoreTable title="PAIR MODE" nameHeader="PLAYERS" rows={pairRows} />
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
