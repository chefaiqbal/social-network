'use client'

import {Sidebar} from '@/pages/feed'

const Profile = () => {
    return (
      <div className="flex h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900">
          <Sidebar></Sidebar>
          <div className="p-7">
            <h1 className="text-4xl	font-bold bg-gradient-to-r from-blue-400 via-purple-500 to-pink-500 bg-clip-text text-transparent">Your portfolio</h1>
            
          </div>        
      </div>
    );
};
  
export default Profile;