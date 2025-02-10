import {useEffect, useState} from "react";
import Draggable from "react-draggable";
import {Button, Drawer} from "@arco-design/web-react";
import {IconMessage} from "@arco-design/web-react/icon";
import {render as amisRender} from "amis";

const FloatingChatGPTButton = () => {
    const [visible, setVisible] = useState(false);
    const [position, setPosition] = useState({x: 3000, y: 3000});


    useEffect(() => {
        // 如果有保存的位置，解析并设置
        const buttonWidth = 30; // 按钮宽度
        const buttonHeight = 30; // 按钮高度
        const maxX = window.innerWidth - buttonWidth - 200; // 右下角的x位置
        const maxY = window.innerHeight - buttonHeight - 100; // 右下角的y位置

        // 从 localStorage 获取上次保存的位置
        const savedPosition = localStorage.getItem("buttonPosition");
        if (savedPosition) {
            const parsedPosition = JSON.parse(savedPosition);
            // 检查位置是否超出范围
            const validX = Math.min(Math.max(parsedPosition.x, 0), maxX);
            const validY = Math.min(Math.max(parsedPosition.y, 0), maxY);
            setPosition({x: validX, y: validY});
        } else {
            // 如果没有保存位置，计算并设置右下角的位置
            const x = maxX;
            const y = maxY;
            setPosition({x, y});
        }
    }, []);

    const handleDrag = (_: any, data: any) => {
        const newPosition = {x: data.x, y: data.y};
        setPosition(newPosition);
        // 每次拖动时，更新位置到 localStorage
        localStorage.setItem("buttonPosition", JSON.stringify(newPosition));
    };

    return (
        <>
            {/* 可拖动按钮 */}
                <Draggable
                    bounds="#root" // 限制在视口内拖动
                    position={position}
                    onDrag={handleDrag}

                >
                    <Button
                        type="primary"
                        shape="circle"
                        icon={<IconMessage/>}
                        style={{
                            position: "fixed",
                            zIndex: 180000, // 确保不被遮挡
                            width: "30px",
                            height: "30px",
                            boxShadow: "0 4px 6px rgba(0, 0, 0, 0.1)",
                            cursor: "grab", // 拖动手势
                        }}
                        onClick={() => setVisible(true)}
                    />
                </Draggable>

                {/* 右侧抽屉 */}
                <Drawer
                    title="问AI"
                    width={600}
                    visible={visible}
                    onCancel={() => setVisible(false)}
                    footer={null}
                    zIndex={180000}
                >
                    {amisRender(
                        {
                            "type": "chatgpt",
                            "url": "/k8s/chat/ws_chatgpt",
                        }
                    )}
                </Drawer>
        </>
    );
};

export default FloatingChatGPTButton;