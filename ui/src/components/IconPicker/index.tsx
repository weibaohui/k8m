import React, { useState, useEffect } from 'react';
import { Modal, Pagination } from 'antd';
import iconOptions from '@/utils/iconOptions';
import './index.module.scss';

interface IconPickerProps {
  visible: boolean;
  onCancel: () => void;
  onSelect: (iconValue: string) => void;
  selectedIcon?: string;
}

const IconPicker: React.FC<IconPickerProps> = ({ visible, onCancel, onSelect, selectedIcon }) => {
  const [currentPage, setCurrentPage] = useState(1);
  const [iconsPerPage] = useState(64); // 8x8网格
  const [displayIcons, setDisplayIcons] = useState<{ value: string }[]>([]);

  // 计算当前页显示的图标
  useEffect(() => {
    const startIndex = (currentPage - 1) * iconsPerPage;
    const endIndex = startIndex + iconsPerPage;
    setDisplayIcons(iconOptions.slice(startIndex, endIndex));
  }, [currentPage, iconsPerPage]);

  // 处理页码变化
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
  };

  // 处理图标选择
  const handleIconSelect = (iconValue: string) => {
    onSelect(iconValue);
    onCancel();
  };

  return (
    <Modal
      title="选择图标"
      visible={visible}
      onCancel={onCancel}
      footer={null}
      width={600}
    >
      <div className="icon-picker-container" style={{ padding: '16px', width: '100%', boxSizing: 'border-box' }}>
        <div 
          className="icon-grid"
          style={{ 
            display: 'grid',
            gridTemplateColumns: 'repeat(8, 1fr)',
            gap: '12px',
            marginBottom: '20px',
            maxHeight: '400px',
            overflowY: 'auto',
            width: '100%',
            boxSizing: 'border-box'
          }}
        >
          {displayIcons.map(icon => (
            <div
              key={icon.value}
              className={`icon-item ${selectedIcon === icon.value ? 'selected' : ''}`}
              onClick={() => handleIconSelect(icon.value)}
              style={{ 
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                width: '48px',
                height: '48px',
                border: '1px solid #eee',
                borderRadius: '4px',
                cursor: 'pointer',
                transition: 'all 0.2s',
                fontSize: '24px',
                backgroundColor: selectedIcon === icon.value ? '#e6f7ff' : '#fff',
                borderColor: selectedIcon === icon.value ? '#1890ff' : '#eee',
                color: selectedIcon === icon.value ? '#1890ff' : 'inherit'
              }}
            >
              <i className={`fa-solid ${icon.value}`}></i>
            </div>
          ))}
        </div>

        <div className="pagination-container" style={{ display: 'flex', justifyContent: 'center', marginTop: '16px' }}>
          <Pagination
            current={currentPage}
            pageSize={iconsPerPage}
            total={iconOptions.length}
            onChange={handlePageChange}
            showSizeChanger={false}
             size="small"
          />
        </div>
      </div>
    </Modal>
  );
};

export default IconPicker;