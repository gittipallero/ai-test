import React from 'react';
import './Slider.css';

interface SliderProps {
    label: string;
    min: number;
    max: number;
    value: number;
    onChange: (value: number) => void;
    className?: string;
}

const Slider: React.FC<SliderProps> = ({ 
    label, 
    min, 
    max, 
    value, 
    onChange,
    className = ''
}) => {
    return (
        <div className={`slider-container ${className}`}>
            <label className="slider-label">
                {label}: {value}
            </label>
            <input 
                type="range" 
                min={min} 
                max={max} 
                value={value} 
                onChange={(e) => onChange(parseInt(e.target.value))}
                className="slider-input"
            />
        </div>
    );
};

export default Slider;
