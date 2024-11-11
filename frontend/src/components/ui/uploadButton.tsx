'use client';

import { useState } from 'react';

// Define props interface with Base64 string
interface UploaderProps {
  onUpload: (base64: string) => void;
}

const Uploader: React.FC<UploaderProps> = ({ onUpload }) => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null); 
  const maxFileSize = 5 * 1024; // 5kb in bytes
  const [error, setError] = useState<string | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);

  const handleUpload = () => {
    if (!selectedFile) {
      setError('Please select a file');
      return;
    }
    if (selectedFile.size > maxFileSize) {
      setError('File size exceeds the maximum limit of 4kb');
      return;
    }
  
    const reader = new FileReader();
    reader.onloadend = () => {
      const base64String = reader.result as string;
      console.log('Base64 string generated:', base64String); // Log the base64 data
      onUpload(base64String); 
      setSelectedFile(null);
      setPreviewUrl(null);
      setError(null);
    };
    reader.readAsDataURL(selectedFile);
  };  

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setSelectedFile(file);
      setPreviewUrl(URL.createObjectURL(file));
    }
  };

  const handleRemove = () => {
    setSelectedFile(null);
    setPreviewUrl(null);
    setError(null);
  };

  return (
    <div className="flex flex-col items-start mt-4">
      <label
        htmlFor="file-upload"
        className="cursor-pointer inline-flex items-center px-3 py-1.5 bg-gray-800/50 text-gray-200 border border-gray-700/50 rounded-lg transition-colors hover:bg-gray-700/50"
      >
        {selectedFile ? selectedFile.name : "Insert Image"}
        <input
          id="file-upload"
          type="file"
          onChange={handleFileChange}
          className="hidden"
          accept="image/png, image/gif, image/jpg"
        />
      </label>

      {selectedFile && (
        <>
          <p className="text-sm text-gray-400 mt-2">
            Selected file: <span className="text-gray-200">{selectedFile.name}</span>
          </p>
          {previewUrl && (
            <img src={previewUrl} alt="Preview" className="mt-2 h-20 w-20 object-cover rounded" />
          )}
          <button
            onClick={handleUpload}
            className="mt-2 px-3 py-1.5 bg-blue-600 text-white rounded-lg"
          >
            Upload
          </button>
          <button
            onClick={handleRemove}
            className="mt-2 px-3 py-1.5 bg-red-600 text-white rounded-lg"
          >
            Remove
          </button>
        </>
      )}

      {error && <p className="text-sm text-red-500 mt-2">{error}</p>}
    </div>
  );
};

export default Uploader;
