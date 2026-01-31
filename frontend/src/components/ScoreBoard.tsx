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

    const renderSingleTable = () => (
        <div className="scoreboard-column">
            <h3 className="column-title">SINGLE PLAYER</h3>
            <div className="scoreboard-table">
                <div className="scoreboard-header">
                    <span className="rank-col">RANK</span>
                    <span className="name-col">PLAYER</span>
                    <span className="score-col">SCORE</span>
                </div>
                {singleScores.length === 0 ? (
                    <div className="scoreboard-empty">NO SCORES</div>
                ) : (
                    singleScores.map((entry, index) => (
                        <div key={`${entry.nickname}-${index}`} className="scoreboard-row">
                            <span className="rank-col">{index + 1}.</span>
                            <span className="name-col">{entry.nickname}</span>
                            <span className="score-col">{entry.score}</span>
                        </div>
                    ))
                )}
            </div>
        </div>
    );

    const renderPairTable = () => (
        <div className="scoreboard-column">
            <h3 className="column-title">PAIR MODE</h3>
            <div className="scoreboard-table">
                <div className="scoreboard-header">
                    <span className="rank-col">RANK</span>
                    <span className="name-col">PLAYERS</span>
                    <span className="score-col">SCORE</span>
                </div>
                {pairScores.length === 0 ? (
                    <div className="scoreboard-empty">NO SCORES</div>
                ) : (
                    pairScores.map((entry, index) => (
                        <div key={`${entry.player1}-${entry.player2}-${index}`} className="scoreboard-row">
                            <span className="rank-col">{index + 1}.</span>
                            <span className="name-col text-small">{entry.player1} & {entry.player2}</span>
                            <span className="score-col">{entry.score}</span>
                        </div>
                    ))
                )}
            </div>
        </div>
    );

    return (
        <div className="scoreboard-container">
            <h2 className="scoreboard-title">*** HIGH SCORES ***</h2>
            
            {loading && <div className="scoreboard-loading">LOADING...</div>}
            {error && <div className="scoreboard-error">{error}</div>}
            
            {!loading && !error && (
                <div className="scoreboards-wrapper">
                    {renderSingleTable()}
                    {renderPairTable()}
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
