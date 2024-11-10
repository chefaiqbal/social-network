'use client'

import { useState } from 'react'
import { motion } from 'framer-motion'

const Uploader = () => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setSelectedFile(file);
    }
  };

  return (
    <div className="flex flex-col items-start	 mt-4" >
      <label
        htmlFor="file-upload"
        className="cursor-pointer inline-flex items-center px-3 py-1.5 bg-gray-800/50 text-gray-200 border border-gray-700/50 rounded-lg transition-colors hover:bg-gray-700/50"
      >
        {selectedFile ? selectedFile.name : "insert image"}
        <input
          id="file-upload"
          type="file"
          onChange={handleFileChange}
          className="hidden"
        />
      </label>

      {selectedFile && (
        <p className="text-sm text-gray-400 mt-2">
          Selected file: <span className="text-gray-200">{selectedFile.name}</span>
        </p>
      )}
    </div>
  );
}

export default Uploader;
