import Link from 'next/link';
import { motion } from 'framer-motion';
import { Home, Users, User, Bell } from 'lucide-react';

const navItems = [
  { href: '/feed', icon: Home, label: 'Feed' },
  { href: '/profile', icon: User, label: 'Profile' },
  { href: '/groups', icon: Users, label: 'Groups' },
  { href: '/follow', icon: Bell, label: 'Follow' },
];

export default function Sidebar() {
  return (
    <div className="bg-white/10 backdrop-blur-lg w-64 shadow-lg border border-gray-700/50 h-[calc(100vh-64px)]">
      <nav className="p-4">
        {navItems.map((item, index) => (
          <motion.div
            key={item.href}
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: index * 0.1 }}
          >
            <Link href={item.href} className="flex items-center px-3 py-2 text-gray-300 hover:bg-gray-800/50 rounded-lg">
              <item.icon className="mr-3" size={20} />
              {item.label}
            </Link>
          </motion.div>
        ))}
      </nav>
    </div>
  );
} 