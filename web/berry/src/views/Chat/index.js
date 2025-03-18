import React, { useEffect, useState } from "react";
import { API } from "../../utils/api";
import "./index.css";

const useIsSmallScreen = () => {
    const [isSmallScreen, setIsSmallScreen] = useState(window.innerWidth <= 768);

    useEffect(() => {
        const handleResize = () => {
            setIsSmallScreen(window.innerWidth <= 768);
        };

        window.addEventListener('resize', handleResize);

        return () => {
            window.removeEventListener('resize', handleResize);
        };
    }, []);

    return isSmallScreen;
};

const Chat = () => {
    const [chatUrl, setChatUrl] = useState("");
    const [loading, setLoading] = useState(true);
    // const isSmallScreen = useIsSmallScreen();

    const loadTokens = async () => {
        try {
            const res = await API.get(`/api/token/`);
            const siteInfo = JSON.parse(localStorage.getItem('siteInfo'));
            if (!siteInfo) {
                console.error("siteInfo not found in localStorage");
                setLoading(false);
                return;
            }
            // const url = `https://like.chatapi.asia/#/?settings={"key":"sk-xxx","url":"https://chat.chatapi.asia"}`;
            const serverAddress = siteInfo.server_address;
            const key = res.data.data[0].key;
            const url = `${siteInfo.chat_link}/#/?settings={"key":"sk-${key}","url":"${serverAddress}"}`;

            setChatUrl(url);
        } catch (error) {
            console.error("Error loading tokens:", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        loadTokens();
    }, []);

    if (loading) {
        return <div className="chat-container">Loading...</div>;
    }

    return (
        <div className="chat-container">
            <iframe
                src={chatUrl}
                style={{ height: '100%', width: '100%', padding: 0, border: 'none' }}
                title="Chat"
            />
        </div>
    );
};

export default Chat;
