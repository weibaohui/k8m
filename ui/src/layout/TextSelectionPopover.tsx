import React, {useState, useEffect} from "react";
import ReactDOM from "react-dom";
import {render as amisRender} from "amis";
import {Card} from "amis-ui";
import Draggable from "react-draggable";
import './TextSelectionPopover.css';


const GlobalTextSelector: React.FC = () => {
    const [selection, setSelection] = useState<{ text: string; x: number; y: number } | null>(null);

    useEffect(() => {
        const handleMouseUp = (event: MouseEvent) => {
            // 检查点击是否发生在卡片内部
            const card = document.querySelector('.selection-card');
            if (card && card.contains(event.target as Node)) {
                return;
            }
            const selectedText = window.getSelection()?.toString().trim();

            if (!selectedText) {
                setSelection(null);
                return;
            }

            setSelection({
                text: selectedText,
                x: event.clientX + window.scrollX, // ✅ 使用鼠标点击的 X 坐标
                y: event.clientY + window.scrollY  // ✅ 使用鼠标点击的 Y 坐标
            });

        };

        document.addEventListener("mouseup", handleMouseUp);
        return () => {
            document.removeEventListener("mouseup", handleMouseUp);
        };
    }, []);

    if (!selection) return null;


    return ReactDOM.createPortal(
        <Draggable handle=".selection-title">
            <div
                className="selection-card"
                style={{
                    position: "absolute",
                    top: selection.y + 5,
                    left: selection.x,
                    zIndex: 100000000,
                    overflow: "auto"
                }}
            >
                <Card style={{width: '50hv', maxWidth: '500px'}}
                      titleClassName="selection-title"
                      title={<>
                          <i className="fas fa-grip-vertical" style={{marginRight: '8px'}}></i>
                          {selection.text.length > 40 ? selection.text.slice(0, 40) + "..." : selection.text}
                      </>}
                >
                    <div style={{
                        display: "flex",
                        flexDirection: "column",
                        alignItems: "center",
                        maxHeight: "50vh",
                        overflow: "auto"
                    }}>
                        {
                            amisRender({
                                "type": "websocketMarkdownViewer",
                                "url": "/k8s/chat/any_selection",
                                "params": {
                                    "question": selection.text
                                }
                            })
                        }
                    </div>
                </Card>


            </div>
        </Draggable>,
        document.body
    );
};

export default GlobalTextSelector;
