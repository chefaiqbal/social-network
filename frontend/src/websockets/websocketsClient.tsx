'use client';

import { useEffect, useState, useRef } from 'react';

interface Likes {
    UserID: number;
    PostID?: number;
    CommentID?: number;
    Like: boolean;
}

interface Notification {
    id: number;          
    ToUserID: number;  
    content: string;    
    FromUserID: number; 
    read: boolean;       
    groupID?: number;    
    createdAt: Date;     
    type: string;
}

interface ChatMessage {
    id?: number;
    sender_id: number;
    recipient_id: number;
    content: string;
    created_at: string;
}

interface GroupMessage {
    id: number;
    groupID: number;
    userID: number;
    content?: string;
    media?: string;
    createdAt: Date;
}

interface WebSocketClientProps {
    onNotification?: (notification: Notification) => void;
    onChatMessage?: (message: ChatMessage) => void;
    onLike?: (like: Likes) => void;
    onGroupMessage?: (message: GroupMessage) => void;
    onUserStatus?: (userId: number, isOnline: boolean) => void;
}

const WebSocketClient = ({
    onNotification,
    onChatMessage,
    onLike,
    onGroupMessage,
    onUserStatus
}: WebSocketClientProps) => {
    const ws = useRef<WebSocket | null>(null);
    const [isConnected, setIsConnected] = useState(false);
    const reconnectTimeoutRef = useRef<NodeJS.Timeout>();

    const connectWebSocket = () => {
        if (ws.current?.readyState !== WebSocket.OPEN) {
            try {
                ws.current = new WebSocket('ws://localhost:8080/ws');

                ws.current.onopen = () => {
                    console.log('WebSocket connection established');
                    setIsConnected(true);
                };

                ws.current.onmessage = (event) => {
                    try {
                        const data = JSON.parse(event.data);
                        console.log('WebSocket message received:', data);

                        switch (data.type) {
                            case 'notification':
                                onNotification?.(data.data);
                                break;
                            case 'chat':
                                onChatMessage?.(data);
                                break;
                            case 'like':
                                onLike?.(data.like);
                                break;
                            case 'groupChat':
                                onGroupMessage?.(data.group_chat);
                                break;
                            case 'user_status':
                                onUserStatus?.(data.user_id, data.is_online);
                                break;
                            default:
                                console.log('Unknown message type:', data.type);
                        }
                    } catch (error) {
                        console.error('Error parsing WebSocket message:', error);
                    }
                };

                ws.current.onerror = (error) => {
                    console.error('WebSocket error:', error);
                    setIsConnected(false);
                };

                ws.current.onclose = (event) => {
                    console.log('WebSocket connection closed:', event.code, event.reason);
                    setIsConnected(false);

                    // Clear any existing reconnection timeout
                    if (reconnectTimeoutRef.current) {
                        clearTimeout(reconnectTimeoutRef.current);
                    }

                    // Only attempt to reconnect if it wasn't a normal closure
                    if (event.code !== 1000) {
                        reconnectTimeoutRef.current = setTimeout(() => {
                            console.log('Attempting to reconnect...');
                            connectWebSocket();
                        }, 3000);
                    }
                };
            } catch (error) {
                console.error('Error creating WebSocket connection:', error);
                setIsConnected(false);
            }
        }
    };

    useEffect(() => {
        connectWebSocket();

        return () => {
            // Clear any pending reconnection attempt
            if (reconnectTimeoutRef.current) {
                clearTimeout(reconnectTimeoutRef.current);
            }

            // Close the WebSocket connection properly
            if (ws.current) {
                const socket = ws.current;
                // Remove all listeners first
                socket.onclose = null;
                socket.onerror = null;
                socket.onmessage = null;
                socket.onopen = null;
                
                // Then close the connection with a normal closure code
                if (socket.readyState === WebSocket.OPEN) {
                    socket.close(1000, 'Component unmounting');
                }
            }
        };
    }, [onNotification, onChatMessage, onLike, onGroupMessage, onUserStatus]);

    // Function to send messages through WebSocket
    const sendMessage = (message: any) => {
        if (ws.current?.readyState === WebSocket.OPEN) {
            try {
                ws.current.send(JSON.stringify(message));
            } catch (error) {
                console.error('Error sending message:', error);
                // If there's an error, try to reconnect
                setIsConnected(false);
                connectWebSocket();
            }
        } else {
            console.error('WebSocket is not connected');
            // Try to reconnect if not connected
            setIsConnected(false);
            connectWebSocket();
        }
    };

    return {
        isConnected,
        sendMessage,
        websocket: ws.current
    };
};

export default WebSocketClient;