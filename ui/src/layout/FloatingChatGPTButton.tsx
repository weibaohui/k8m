import {useEffect, useState, useRef} from "react";
import Draggable from "react-draggable";
import {render as amisRender} from "amis";
import {Button, Drawer} from "antd";
import {OpenAIFilled} from "@ant-design/icons";

const FloatingChatGPTButton = () => {
    const [visible, setVisible] = useState(false);
    const [position, setPosition] = useState({x: 3000, y: 3000});
    const [isDragging, setIsDragging] = useState(false); // 拖拽状态
    const startPositionRef = useRef({x: 0, y: 0}); // 鼠标点击的初始位置

    useEffect(() => {
        const buttonWidth = 30;
        const buttonHeight = 30;
        const maxX = window.innerWidth - buttonWidth - 200 - 60;
        const maxY = window.innerHeight - buttonHeight - 100;

        const savedPosition = localStorage.getItem("buttonPosition");
        if (savedPosition) {
            const parsedPosition = JSON.parse(savedPosition);
            const validX = Math.min(Math.max(parsedPosition.x, 0), maxX);
            const validY = Math.min(Math.max(parsedPosition.y, 0), maxY);
            setPosition({x: validX, y: validY});
        } else {
            setPosition({x: maxX, y: maxY});
        }
    }, []);

    const handleDrag = (_: any, data: any) => {
        setPosition({x: data.x, y: data.y});
        localStorage.setItem("buttonPosition", JSON.stringify({x: data.x, y: data.y}));
    };

    const handleStart = (e: any) => {
        startPositionRef.current = {x: e.clientX, y: e.clientY};
        setIsDragging(true); // 开始拖动
    };

    const handleStop = (e: any) => {
        const endPosition = {x: e.clientX, y: e.clientY};
        const distance = Math.sqrt(
            Math.pow(endPosition.x - startPositionRef.current.x, 2) +
            Math.pow(endPosition.y - startPositionRef.current.y, 2)
        );
        if (distance < 1) { // 如果拖动距离小于5px，认为是点击
            setIsDragging(false);
        } else {
            setTimeout(() => setIsDragging(false), 200); // 结束拖动，延迟一点避免误触发点击
        }
    };

    return (
        <>
            {/* 可拖动按钮 */}
            <Draggable
                bounds="#root"
                position={position}
                onStart={handleStart} // 开始拖动时记录初始位置
                onDrag={handleDrag}
                onStop={handleStop} // 结束拖动时计算距离
            >
                <Button
                    type="primary"
                    shape="circle"
                    icon={<OpenAIFilled/>}
                    style={{
                        position: "fixed",
                        zIndex: 180000,
                        width: "30px",
                        height: "30px",
                        boxShadow: "0 4px 6px rgba(0, 0, 0, 0.1)",
                        cursor: "grab",
                    }}
                    onClick={() => {
                        if (!isDragging) setVisible(true); // 只有非拖动状态下才触发点击
                    }}
                />
            </Draggable>

            {/* 右侧抽屉 */}
            <Drawer
                title="问AI"
                width={600}
                onClose={() => setVisible(false)}
                footer={null}
                zIndex={180000}
                open={visible}
                maskClosable={false}
            >
                {amisRender({
                    type: "chatgpt",
                    url: "/ai/chat/ws_chatgpt",
                })}
            </Drawer>
        </>
    );
};

export default FloatingChatGPTButton;