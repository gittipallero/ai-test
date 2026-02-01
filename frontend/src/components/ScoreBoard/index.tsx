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
}

const ScoreBoard: React.FC<ScoreBoardProps> = ({ onBack }) => {
    const [singleScores, setSingleScores] = useState<ScoreEntry[]>([]);
    const [pairScores, setPairScores] = useState<PairScoreEntry[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        Promise.all([
            fetch('/api/scoreboard').then(res => res.json()),
            fetch('/api/scoreboard/pair').then(res => res.json())
        ])
            .then(([singleData, pairData]) => {
                setSingleScores(singleData || []);
                setPairScores(pairData || []);
                setLoading(false);
            })
            .catch(err => {
                console.error('Scoreboard error:', err);
                setError('Failed to load scoreboards');
                setLoading(false);
            });
    }, []);

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
