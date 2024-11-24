'use client';

import { useState } from 'react';
import { motion } from 'framer-motion';
import Input from '../ui/Input';
import Button from '../ui/Button';
import Uploader from '../ui/uploadButton';
import { api } from '@/lib/api';
import { useRouter } from 'next/navigation';
import { toast } from 'react-hot-toast';
import { useAuth } from '@/hooks/useAuth';
import { User } from 'lucide-react';

export default function RegisterForm() {
  const router = useRouter();
  const setUser = useAuth((state) => state.setUser);
  const [isLoading, setIsLoading] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [formData, setFormData] = useState({
    email: '',
    username: '',
    password: '',
    firstName: '',
    lastName: '',
    dateOfBirth: '',
    aboutMe: '',
    avatar: '', // This will store the base64 string
  });
  const [avatarPreview, setAvatarPreview] = useState<string | null>(null);

  const validateForm = () => {
    const newErrors: Record<string, string> = {};
    
    if (!formData.email.includes('@')) {
      newErrors.email = 'Please enter a valid email';
    }
    if (formData.password.length < 6) {
      newErrors.password = 'Password must be at least 6 characters';
    }
    if (!formData.username) {
      newErrors.username = 'Username is required';
    }
    if (!formData.firstName) {
      newErrors.firstName = 'First name is required';
    }
    if (!formData.lastName) {
      newErrors.lastName = 'Last name is required';
    }
    if (!formData.dateOfBirth) {
      newErrors.dateOfBirth = 'Date of birth is required';
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleAvatarUpload = (base64: string) => {
    setFormData(prev => ({ ...prev, avatar: base64 }));
    setAvatarPreview(base64);
    setErrors(prev => ({ ...prev, avatar: '' }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;
    
    setIsLoading(true);
    try {
      const userData = await api.register({
        email: formData.email,
        username: formData.username,
        password: formData.password,
        first_name: formData.firstName,
        last_name: formData.lastName,
        date_of_birth: formData.dateOfBirth,
        about_me: formData.aboutMe,
        avatar: formData.avatar, // Send the base64 string
      });
      
      setUser(userData);
      toast.success('Registration successful!');
      router.push('/feed');
    } catch (error) {
      console.error('Registration error:', error);
      toast.error(error instanceof Error ? error.message : 'Registration failed');
      setErrors({ submit: 'Registration failed. Please try again.' });
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <motion.form
      onSubmit={handleSubmit}
      className="space-y-6"
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
    >
      <div className="space-y-6">
        {/* Avatar Upload Section */}
        <div className="flex flex-col items-center space-y-4">
          <div className="relative w-24 h-24 rounded-full overflow-hidden border-2 border-gray-700/50 bg-gray-800/50">
            {avatarPreview ? (
              <img 
                src={avatarPreview} 
                alt="Avatar preview" 
                className="w-full h-full object-cover"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center">
                <User size={40} className="text-gray-400" />
              </div>
            )}
          </div>
          <Uploader onUpload={handleAvatarUpload} />
          {errors.avatar && (
            <p className="text-sm text-red-400">{errors.avatar}</p>
          )}
        </div>

        <div className="grid grid-cols-2 gap-4">
          <Input
            type="text"
            label="First Name"
            value={formData.firstName}
            onChange={(e) => setFormData({ ...formData, firstName: e.target.value })}
            error={errors.firstName}
            required
          />
          <Input
            type="text"
            label="Last Name"
            value={formData.lastName}
            onChange={(e) => setFormData({ ...formData, lastName: e.target.value })}
            error={errors.lastName}
            required
          />
        </div>

        <Input
          type="text"
          label="Username"
          value={formData.username}
          onChange={(e) => setFormData({ ...formData, username: e.target.value })}
          error={errors.username}
          required
        />

        <Input
          type="email"
          label="Email"
          value={formData.email}
          onChange={(e) => setFormData({ ...formData, email: e.target.value })}
          error={errors.email}
          required
        />

        <Input
          type="password"
          label="Password"
          value={formData.password}
          onChange={(e) => setFormData({ ...formData, password: e.target.value })}
          error={errors.password}
          required
        />

        <Input
          type="date"
          label="Date of Birth"
          value={formData.dateOfBirth}
          onChange={(e) => setFormData({ ...formData, dateOfBirth: e.target.value })}
          error={errors.dateOfBirth}
          required
        />

        <Input
          type="text"
          label="About Me"
          value={formData.aboutMe}
          onChange={(e) => setFormData({ ...formData, aboutMe: e.target.value })}
          placeholder="Tell us about yourself..."
          error={errors.aboutMe}
          required
        />
      </div>

      {errors.submit && (
        <p className="text-sm text-red-400 text-center">{errors.submit}</p>
      )}

      <Button
        type="submit"
        fullWidth
        variant="gradient"
        isLoading={isLoading}
      >
        Create Account
      </Button>
    </motion.form>
  );
} 