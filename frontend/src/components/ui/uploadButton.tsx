'use client';

import { useState, useCallback } from 'react';
import { X, Upload, Image as ImageIcon } from 'lucide-react';

interface UploaderProps {
  onUpload: (base64: string, mediaType: string) => void;
  buttonText?: string;
  acceptedTypes?: string;
}

const Uploader: React.FC<UploaderProps> = ({ 
  onUpload, 
  buttonText = "Insert Image",
  acceptedTypes = "image/jpeg,image/png,image/gif,image/webp"
}) => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  // Validate file type and size
  const validateFile = (file: File): boolean => {
    const validTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
    const maxSize = 5 * 1024 * 1024; // 5MB

    if (!validTypes.includes(file.type)) {
      setError('Invalid file type. Please upload a JPEG, PNG, GIF, or WebP image.');
      return false;
    }

    if (file.size > maxSize) {
      setError('File size too large. Maximum size is 5MB.');
      return false;
    }

    return true;
  };

  // Compress image if needed (skip for GIFs)
  const compressImage = async (file: File): Promise<Blob> => {
    // Don't compress GIFs to preserve animation
    if (file.type === 'image/gif') {
      return file;
    }

    return new Promise((resolve, reject) => {
      const img = new Image();
      img.src = URL.createObjectURL(file);

      img.onload = () => {
        const canvas = document.createElement('canvas');
        let width = img.width;
        let height = img.height;

        // Maximum dimensions
        const MAX_WIDTH = 1200;
        const MAX_HEIGHT = 1200;

        if (width > height) {
          if (width > MAX_WIDTH) {
            height = Math.round((height * MAX_WIDTH) / width);
            width = MAX_WIDTH;
          }
        } else {
          if (height > MAX_HEIGHT) {
            width = Math.round((width * MAX_HEIGHT) / height);
            height = MAX_HEIGHT;
          }
        }

        canvas.width = width;
        canvas.height = height;

        const ctx = canvas.getContext('2d');
        if (!ctx) {
          reject(new Error('Failed to get canvas context'));
          return;
        }

        ctx.drawImage(img, 0, 0, width, height);

        canvas.toBlob(
          (blob) => {
            if (blob) {
              resolve(blob);
            } else {
              reject(new Error('Failed to compress image'));
            }
          },
          'image/jpeg',
          0.8 // Quality
        );
      };

      img.onerror = () => {
        reject(new Error('Failed to load image'));
      };
    });
  };

  const handleUpload = async () => {
    if (!selectedFile) {
      setError('Please select a file');
      return;
    }

    try {
      setIsLoading(true);
      setError(null);

      // Process the image (compress if not GIF)
      const processedBlob = await compressImage(selectedFile);
      
      // Convert to base64
      const reader = new FileReader();
      reader.onloadend = () => {
        const base64String = reader.result as string;
        onUpload(base64String, selectedFile.type); // Pass both base64 and media type
        handleRemove(); // Clear the form after successful upload
      };
      reader.readAsDataURL(processedBlob);
    } catch (err) {
      setError('Failed to process image. Please try again.');
      console.error('Upload error:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleFileChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      if (validateFile(file)) {
        setSelectedFile(file);
        setPreviewUrl(URL.createObjectURL(file));
        setError(null);
      }
    }
  }, []);

  const handleRemove = useCallback(() => {
    setSelectedFile(null);
    setPreviewUrl(null);
    setError(null);
    const input = document.getElementById('file-upload') as HTMLInputElement;
    if (input) input.value = '';
  }, []);

  return (
    <div className="flex flex-col items-start mt-4">
      <div className="w-full">
        {!selectedFile ? (
          <label
            htmlFor="file-upload"
            className="cursor-pointer inline-flex items-center gap-2 px-3 py-1.5 bg-gray-800/50 text-gray-200 border border-gray-700/50 rounded-lg transition-colors hover:bg-gray-700/50"
          >
            <ImageIcon size={18} />
            {buttonText}
            <input
              id="file-upload"
              type="file"
              onChange={handleFileChange}
              className="hidden"
              accept={acceptedTypes}
            />
          </label>
        ) : (
          <div className="space-y-3">
            <div className="flex items-start gap-4">
              {previewUrl && (
                <div className="relative group">
                  <img 
                    src={previewUrl} 
                    alt="Preview" 
                    className="h-20 w-20 object-cover rounded-lg border border-gray-700/50"
                    style={{
                      objectFit: selectedFile.type === 'image/gif' ? 'contain' : 'cover'
                    }}
                  />
                  <div className="absolute inset-0 bg-black/50 opacity-0 group-hover:opacity-100 transition-opacity rounded-lg flex items-center justify-center">
                    <button
                      onClick={handleRemove}
                      className="text-white p-1 hover:text-red-400"
                    >
                      <X size={20} />
                    </button>
                  </div>
                </div>
              )}
              <div className="flex-1 space-y-2">
                <p className="text-sm text-gray-400 truncate">
                  {selectedFile.name}
                </p>
                <div className="flex gap-2">
                  <button
                    onClick={handleUpload}
                    disabled={isLoading}
                    className="px-3 py-1.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2"
                  >
                    <Upload size={18} />
                    {isLoading ? 'Processing...' : 'Upload'}
                  </button>
                  <button
                    onClick={handleRemove}
                    className="px-3 py-1.5 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
                  >
                    Remove
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {error && (
        <p className="text-sm text-red-500 mt-2 flex items-center gap-1">
          <X size={16} />
          {error}
        </p>
      )}
    </div>
  );
};

export default Uploader;
