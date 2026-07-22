import React, { useEffect } from 'react';
import { DatakeysWidget } from './DatakeysWidget';
import type { DatakeysWidgetProps } from './types';

interface ModalProps extends DatakeysWidgetProps {
  isOpen: boolean;
}

export function DatakeysModal({ isOpen, ...props }: ModalProps) {
  useEffect(() => {
    document.body.style.overflow = isOpen ? 'hidden' : '';
    return () => {
      document.body.style.overflow = '';
    };
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div
      style={{
        position: 'fixed',
        inset: 0,
        background: 'rgba(0,0,0,0.5)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        zIndex: 9999,
        padding: '20px',
      }}
    >
      <DatakeysWidget {...props} onClose={props.onClose} />
    </div>
  );
}
