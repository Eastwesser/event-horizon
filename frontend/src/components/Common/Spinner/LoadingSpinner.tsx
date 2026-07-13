// src/components/common/LoadingSpinner.tsx
import React from 'react';
import './LoadingSpinner.scss';

const LoadingSpinner: React.FC = () => {
  return (
    <div className="loading-spinner">
      <div className="spinner"></div>
      <p>Загрузка...</p>
    </div>
  );
};

export default LoadingSpinner;