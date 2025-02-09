import {useState} from "react";
import Draggable from "react-draggable";
import {Button, Drawer} from "@arco-design/web-react";
import {IconMessage} from "@arco-design/web-react/icon";
import {render as amisRender} from "amis";

const FloatingActionButton = () => {
    const [visible, setVisible] = useState(false);
    const [position, setPosition] = useState({x: 20, y: 20}); // 初始位置

    const handleDrag = (_: any, data: any) => {
        setPosition({x: data.x, y: data.y});
    };

    return (
        <>
            {/* 可拖动按钮 */}
            <Draggable
                bounds="parent" // 限制在视口内拖动
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
                        width: "50px",
                        height: "50px",
                        boxShadow: "0 4px 6px rgba(0, 0, 0, 0.1)",
                        cursor: "grab", // 拖动手势
                    }}
                    onClick={() => setVisible(true)}
                />
            </Draggable>

            {/* 右侧抽屉 */}
            <Drawer
                title="Chat窗口"
                width={400}
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

export default FloatingActionButton;