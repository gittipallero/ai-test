import React, { useEffect, useState } from 'react';
import GameButton from './GameButton';
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

interface ScoreTableRow {
    key: string;
    rank: number;
    name: string;
    score: number;
    nameClassName?: string;
}

interface ScoreTableProps {
    title: string;
    nameHeader: string;
    rows: ScoreTableRow[];
}

const ScoreTable: React.FC<ScoreTableProps> = ({ title, nameHeader, rows }) => (
    <div className="scoreboard-column">
        <h3 className="column-title">{title}</h3>
        <div className="scoreboard-table">
            <div className="scoreboard-header">
                <span className="rank-col">RANK</span>
                <span className="name-col">{nameHeader}</span>
                <span className="score-col">SCORE</span>
            </div>
            {rows.length === 0 ? (
                <div className="scoreboard-empty">NO SCORES</div>
            ) : (
                rows.map((row) => (
                    <div key={row.key} className="scoreboard-row">
                        <span className="rank-col">{row.rank}.</span>
                        <span className={['name-col', row.nameClassName].filter(Boolean).join(' ')}>{row.name}</span>
                        <span className="score-col">{row.score}</span>
                    </div>
                ))
            )}
        </div>
    </div>
);

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
