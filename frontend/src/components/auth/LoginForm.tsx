'use client';

import { useState } from 'react';
import { motion } from 'framer-motion';
import { useRouter } from 'next/navigation';
import Input from '../ui/Input';
import Button from '../ui/Button';
import { toast } from 'react-hot-toast';
import { api } from '@/lib/api';
import { useAuth } from '@/hooks/useAuth';

export default function LoginForm() {
  const router = useRouter();
  const setUser = useAuth((state) => state.setUser);
  const [isLoading, setIsLoading] = useState(false);
  const [formData, setFormData] = useState({
    email: '',
    password: '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      const userData = await api.login(formData.email, formData.password);
      setUser(userData);
      toast.success('Login successful!');
      router.push('/feed');
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Login failed';
      toast.error(errorMessage);
      console.error('Login error:', error);
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
      <Input
        type="email"
        label="Email"
        value={formData.email}
        onChange={(e) => setFormData({ ...formData, email: e.target.value })}
        required
        placeholder="Enter your email"
      />
      <Input
        type="password"
        label="Password"
        value={formData.password}
        onChange={(e) => setFormData({ ...formData, password: e.target.value })}
        required
        placeholder="Enter your password"
      />
      <Button
        type="submit"
        fullWidth
        variant="gradient"
        isLoading={isLoading}
      >
        Sign In
      </Button>
    </motion.form>
  );
} 