import { create } from 'zustand';

interface User {
  id: number;
  username: string;
  // add other user fields as needed
}

interface AuthState {
  user: User | null;
  setUser: (user: User | null) => void;
}

export const useAuth = create<AuthState>((set) => ({
  user: null,
  setUser: (user) => set({ user }),
})); 