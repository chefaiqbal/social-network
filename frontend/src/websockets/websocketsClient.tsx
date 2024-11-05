'use client';

import { useEffect, useState } from 'react';

const WebSocketClient = () => {

    const [notifications, setNotifications] = useState<Notification[]>([]);
    const [chats, setChats] = useState<ChatMessage[]>([]);
    const [groupChat, setGroupChat] = useState<GroupMessage[]>([]);
    const [likes, setLikes] = useState<Likes[]>([]);

    useEffect(() => {
        const ws = new WebSocket('ws://localhost:8080/ws');

        ws.onopen = () => {
            console.log('WebSocket connection established');
        };

        ws.onmessage = (event) => {
            const data = JSON.parse(event.data);
            switch (data.type) {
                case 'notification':
                    setNotifications((prevNotifications) => [...prevNotifications, data.notification]);
                    break;
                case 'chat':
                    setChats((prevChats) => [...prevChats, data.chat]);
                    break;
                case 'like':
                    setLikes((prevLikes) => [...prevLikes, data.like]);
                    break;
                case 'groupChat':
                    setGroupChat((prevGroupChat) => [...prevGroupChat, data.group_chat]);
                    break;
                default:
                    console.log('Unknown message type:', data.type);
            }
        };

        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        return () => {
            ws.close();
            console.log('WebSocket connection closed');
        };
    }, []); 
};

export default WebSocketClient;


interface Likes {
    UserID : number;
    PostID? : number;
    CommentID? : number;
    Like : boolean;
}

interface Notification {
    id: number;          
    ToUserID: number;  
    content: string;    
    FromUserID: number; 
    read: boolean;       
    groupID?: number;    
    createdAt: Date;     
}


interface ChatMessage {
    senderID: number;
    recipientID: number;
    content?: string;
    userName?: string;
    createdAt: Date;
}

interface GroupMessage {
    id: number;
    groupID: number;
    userID: number;
    content?: string;
    media?: string;
    createdAt: Date;
}