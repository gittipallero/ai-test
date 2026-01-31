import React from 'react';

export interface ScoreTableRow {
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
                        <span className={['name-col', row.nameClassName].filter(Boolean).join(' ')}>
                            {row.name}
                        </span>
                        <span className="score-col">{row.score}</span>
                    </div>
                ))
            )}
        </div>
    </div>
);

export default ScoreTable;
