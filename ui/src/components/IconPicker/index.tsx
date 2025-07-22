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
  const totalPages = Math.ceil(iconOptions.length / iconsPerPage);

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
      <div className="icon-picker-container">
        <div className="icon-grid">
          {displayIcons.map(icon => (
            <div
              key={icon.value}
              className={`icon-item ${selectedIcon === icon.value ? 'selected' : ''}`}
              onClick={() => handleIconSelect(icon.value)}
            >
              <i className={`fa-solid ${icon.value}`}></i>
            </div>
          ))}
        </div>

        <div className="pagination-container">
          <Pagination
            current={currentPage}
            pageSize={iconsPerPage}
            total={iconOptions.length}
            onChange={handlePageChange}
            showSizeChanger={false}
            showQuickJumper
          />
        </div>
      </div>
    </Modal>
  );
};

export default IconPicker;