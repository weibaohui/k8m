import React, {useState, useEffect} from "react";
import ReactDOM from "react-dom";
import {render as amisRender} from "amis";
import {Card} from "amis-ui";


const GlobalTextSelector: React.FC = () => {
    const [selection, setSelection] = useState<{ text: string; x: number; y: number } | null>(null);

    useEffect(() => {
        const handleMouseUp = (event: MouseEvent) => {
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
        <div
            style={{
                position: "absolute",
                top: selection.y + 5,
                left: selection.x,
                zIndex: 100000000,
            }}
        >
            <Card style={{width: '50hv', maxWidth: '500px'}}
                  title={selection.text.length > 40 ? selection.text.slice(0, 40) + "..." : selection.text}
            >
                <div style={{display: "flex", flexDirection: "column", alignItems: "center"}}>
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


        </div>,
        document.body
    );
};

export default GlobalTextSelector;
