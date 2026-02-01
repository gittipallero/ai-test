import React from 'react';
import './ArrowButton.css';

interface ArrowButtonProps {
    onTrigger: () => void;
    label: string;
    children: React.ReactNode;
    className?: string; // For positioning (e.g., up, down, left, right)
}

const ArrowButton: React.FC<ArrowButtonProps> = ({ 
    onTrigger, 
    label, 
    children, 
    className = '' 
}) => {
    
    const handleTouchStart = (e: React.TouchEvent) => {
        e.preventDefault();
        onTrigger();
    };

    const handleMouseDown = (e: React.MouseEvent) => {
        e.preventDefault();
        onTrigger();
    };

    return (
        <button 
            className={`arrow-btn ${className}`}
            onTouchStart={handleTouchStart}
            onMouseDown={handleMouseDown}
            aria-label={label}
        >
            {children}
        </button>
    );
};

export default ArrowButton;
