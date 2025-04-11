import React, {useState, useEffect} from "react";
import ReactDOM from "react-dom";
import {render as amisRender} from "amis";
import {Card} from "amis-ui";
import Draggable from "react-draggable";
import './TextSelectionPopover.css';
import {fetcher} from "@/components/Amis/fetcher";

const GlobalTextSelector: React.FC = () => {
    const [selection, setSelection] = useState<{ text: string; x: number; y: number } | null>(null);
    const [isFullscreen, setIsFullscreen] = useState(false);
    const [isEnabled, setIsEnabled] = useState(false);

    useEffect(() => {
        // 从后端获取配置
        fetcher({
            url: '/params/config/AnySelect',
            method: 'get'
        })
            .then(response => {
                //@ts-ignore
                setIsEnabled(response.data?.data === 'true');
            })
            .catch(error => {
                console.error('Error fetching AnySelect config:', error);
                setIsEnabled(false);
            });
    }, []);

    useEffect(() => {
        if (!isEnabled) return;

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
                x: event.clientX + window.scrollX,
                y: event.clientY + window.scrollY
            });
        };

        document.addEventListener("mouseup", handleMouseUp);
        return () => {
            document.removeEventListener("mouseup", handleMouseUp);
        };
    }, [isEnabled]);

    if (!selection || !isEnabled) return null;

    return ReactDOM.createPortal(
        <Draggable handle=".selection-title" disabled={isFullscreen}>
            <div
                className="selection-card"
                style={{
                    position: "absolute",
                    top: isFullscreen ? 0 : selection.y + 5,
                    left: isFullscreen ? 0 : selection.x,
                    zIndex: 100000000,
                    width: isFullscreen ? "100vw" : "550px",
                    height: isFullscreen ? "100vh" : "auto",
                    overflow: "auto",
                    transition: "all 0.1s ease"
                }}
            >
                <Card
                    titleClassName="selection-title"
                    title={<>
                        <i className="fas fa-grip-vertical"
                           style={{marginRight: '8px', visibility: isFullscreen ? 'hidden' : 'visible'}}></i>
                        {selection.text.length > 40 ? selection.text.slice(0, 40) + "..." : selection.text}
                        &nbsp;&nbsp;

                        <i
                            className={`fas ${isFullscreen ? 'fa-compress' : 'fa-expand-arrows-alt'}`}
                            style={{
                                marginLeft: 'auto',
                                cursor: 'pointer',
                                fontSize: '14px'
                            }}
                            onClick={() => setIsFullscreen(!isFullscreen)}
                        />

                    </>}
                    style={{
                        height: isFullscreen ? '100%' : 'auto'
                    }}
                >
                    <div style={{
                        display: "flex",
                        flexDirection: "column",
                        alignItems: "center",
                        maxHeight: isFullscreen ? "calc(100vh - 60px)" : "50vh",
                        overflow: "auto"
                    }}>
                        {
                            amisRender({
                                "type": "websocketMarkdownViewer",
                                "url": "/ai/chat/any_selection",
                                "params": {
                                    "question": selection.text
                                },
                                "width": isFullscreen ? "100%" : "500px"
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
