'use client';

import { useState } from 'react';
import LoginForm from '@/components/auth/LoginForm';
import RegisterForm from '@/components/auth/RegisterForm';
import { motion, AnimatePresence } from 'framer-motion';

export default function Home() {
  const [isLogin, setIsLogin] = useState(true);

  return (
    <main className="min-h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 flex items-center justify-center p-4">
      <div className="w-full max-w-6xl flex flex-col md:flex-row gap-8 items-center">
        {/* Left side - Branding */}
        <motion.div 
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.5 }}
          className="flex-1 text-center md:text-left text-white"
        >
          <h1 className="text-5xl md:text-6xl font-bold mb-6">
            <span className="bg-gradient-to-r from-blue-400 via-purple-500 to-pink-500 bg-clip-text text-transparent">
              Reboot Network
            </span>
          </h1>
          <p className="text-xl md:text-2xl text-gray-300 mb-8 leading-relaxed">
            Connect with fellow tech enthusiasts in the Reboot community.
          </p>
          <div className="hidden md:block">
            <div className="bg-gray-800/50 p-8 rounded-xl backdrop-blur-sm border border-gray-700/50">
              <h2 className="text-2xl font-semibold mb-6 text-blue-400">Why Join Us?</h2>
              <ul className="space-y-4">
                <motion.li 
                  className="flex items-center gap-3"
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.2 }}
                >
                  <span className="text-green-400 text-xl">✓</span>
                  <span className="text-lg">Build your tech network</span>
                </motion.li>
                <motion.li 
                  className="flex items-center gap-3"
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.3 }}
                >
                  <span className="text-green-400 text-xl">✓</span>
                  <span className="text-lg">Share your learning journey</span>
                </motion.li>
                <motion.li 
                  className="flex items-center gap-3"
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: 0.4 }}
                >
                  <span className="text-green-400 text-xl">✓</span>
                  <span className="text-lg">Join study groups & events</span>
                </motion.li>
              </ul>
            </div>
          </div>
        </motion.div>

        {/* Right side - Auth Forms */}
        <motion.div 
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.5 }}
          className="flex-1 w-full max-w-md"
        >
          <div className="bg-white/10 backdrop-blur-lg p-8 rounded-2xl shadow-xl border border-gray-700/50">
            <div className="flex justify-center mb-8">
              <div className="inline-flex rounded-lg bg-gray-800 p-1">
                <button
                  onClick={() => setIsLogin(true)}
                  className={`px-6 py-2.5 rounded-lg transition-all duration-300 ${
                    isLogin
                      ? 'bg-blue-500 text-white shadow-lg'
                      : 'text-gray-400 hover:text-white'
                  }`}
                >
                  Login
                </button>
                <button
                  onClick={() => setIsLogin(false)}
                  className={`px-6 py-2.5 rounded-lg transition-all duration-300 ${
                    !isLogin
                      ? 'bg-blue-500 text-white shadow-lg'
                      : 'text-gray-400 hover:text-white'
                  }`}
                >
                  Register
                </button>
              </div>
            </div>
            <AnimatePresence mode="wait">
              <motion.div
                key={isLogin ? 'login' : 'register'}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -10 }}
                transition={{ duration: 0.2 }}
              >
                {isLogin ? <LoginForm /> : <RegisterForm />}
              </motion.div>
            </AnimatePresence>
          </div>
        </motion.div>
      </div>
    </main>
  );
} 