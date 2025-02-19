import React from 'react';
import { InputNumber, Switch, DatePicker } from 'antd';

interface LogOptionsProps {
    tailLines: number;
    follow: boolean;
    timestamps: boolean;
    previous: boolean;
    sinceTime?: string;
    onTailLinesChange: (value: number) => void;
    onFollowChange: (value: boolean) => void;
    onTimestampsChange: (value: boolean) => void;
    onPreviousChange: (value: boolean) => void;
    onSinceTimeChange: (value: string | undefined) => void;
}

const LogOptionsComponent: React.FC<LogOptionsProps> = ({
    tailLines,
    follow,
    timestamps,
    previous,
    onTailLinesChange,
    onFollowChange,
    onTimestampsChange,
    onPreviousChange,
    onSinceTimeChange
}) => {
    return (
        <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
            <InputNumber
                value={tailLines}
                onChange={(value) => onTailLinesChange(value ?? 100)}
                min={0}
                prefix="行数"
                placeholder="显示行数"
                style={{ width: "120px" }}
            />
            <Switch
                checked={follow}
                onChange={onFollowChange}
                checkedChildren="实时"
                unCheckedChildren="实时"
            />
            <Switch
                checked={timestamps}
                onChange={onTimestampsChange}
                checkedChildren="时间戳"
                unCheckedChildren="时间戳"
            />
            <Switch
                checked={previous}
                onChange={onPreviousChange}
                checkedChildren="上一个"
                unCheckedChildren="上一个"
            />
            <DatePicker
                showTime
                format="YYYY-MM-DD HH:mm:ss"
                onChange={(date) => onSinceTimeChange(date?.format('YYYY-MM-DD HH:mm:ss'))}
                placeholder="选择开始时间"
            />
        </div>
    );
};

export default LogOptionsComponent;