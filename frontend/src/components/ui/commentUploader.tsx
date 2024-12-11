import React, { useRef } from 'react';

interface CommentUploaderProps {
    onUpload: (base64: string) => void;
    buttonText?: string;
}

export function CommentUploader({ onUpload, buttonText = "Add Image" }: CommentUploaderProps) {
    const fileInputRef = useRef<HTMLInputElement>(null);

    const handleClick = (e: React.MouseEvent) => {
        e.preventDefault();
        e.stopPropagation();
        fileInputRef.current?.click();
    };

    const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
        e.preventDefault();
        e.stopPropagation();
        
        const file = e.target.files?.[0];
        if (!file) return;

        try {
            const reader = new FileReader();
            reader.onloadend = () => {
                const base64 = reader.result as string;
                onUpload(base64);
            };
            reader.readAsDataURL(file);
        } catch (error) {
            console.error('Error reading file:', error);
        }
    };

    return (
        <div onClick={(e) => e.stopPropagation()}>
            <input
                type="file"
                ref={fileInputRef}
                onChange={handleFileChange}
                accept="image/*"
                style={{ display: 'none' }}
                onClick={(e) => e.stopPropagation()}
            />
            <button
                type="button"
                onClick={handleClick}
                className="px-4 py-2 text-sm text-blue-400 hover:text-blue-300 border border-blue-400/50 rounded-full hover:bg-blue-400/10 transition-colors"
            >
                {buttonText}
            </button>
        </div>
    );
}

export default CommentUploader; 