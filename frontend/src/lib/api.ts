interface RegisterData {
  email: string;
  username: string;
  password: string;
  first_name: string;
  last_name: string;
  date_of_birth: string;
  about_me: string;
  avatar: string;
  nickName: string;
}

interface LoginResponse {
  id: number;
  username: string;
  email: string;
}

const BASE_URL = 'http://localhost:8080';

export const api = {
  async register(data: RegisterData): Promise<LoginResponse> {
    const response = await fetch(`${BASE_URL}/register`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      const errorData = await response.text();
      throw new Error(errorData || 'Registration failed');
    }

    return response.json();
  },

  async login(email: string, password: string): Promise<LoginResponse> {
    const response = await fetch(`${BASE_URL}/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      credentials: 'include',
      body: JSON.stringify({ email, password }),
    });

    if (!response.ok) {
      const errorData = await response.text();
      throw new Error(errorData || 'Login failed');
    }

    return response.json();
  },

  async logout(): Promise<void> {
    const response = await fetch(`${BASE_URL}/logout`, {
      method: 'POST',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Logout failed');
    }
  },
}; 