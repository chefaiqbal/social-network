interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  fullWidth?: boolean;
  variant?: 'primary' | 'secondary' | 'gradient';
  children: React.ReactNode;
  isLoading?: boolean;
}

export default function Button({
  fullWidth,
  variant = 'primary',
  children,
  isLoading,
  ...props
}: ButtonProps) {
  const baseStyles = "relative px-6 py-2.5 rounded-lg font-medium transition-all duration-200 ease-in-out flex items-center justify-center";
  
  const variants = {
    primary: "bg-blue-500 hover:bg-blue-600 text-white shadow-lg shadow-blue-500/25",
    secondary: "bg-gray-700 hover:bg-gray-600 text-white",
    gradient: "bg-gradient-to-r from-blue-500 via-indigo-500 to-purple-500 hover:from-blue-600 hover:via-indigo-600 hover:to-purple-600 text-white shadow-lg shadow-purple-500/25"
  };

  return (
    <button
      {...props}
      disabled={isLoading || props.disabled}
      className={`
        ${baseStyles}
        ${variants[variant]}
        ${fullWidth ? 'w-full' : ''}
        ${isLoading || props.disabled ? 'opacity-50 cursor-not-allowed' : ''}
        group
      `}
    >
      {isLoading ? (
        <svg 
          className="animate-spin h-5 w-5 text-white" 
          xmlns="http://www.w3.org/2000/svg" 
          fill="none" 
          viewBox="0 0 24 24"
        >
          <circle 
            className="opacity-25" 
            cx="12" 
            cy="12" 
            r="10" 
            stroke="currentColor" 
            strokeWidth="4"
          />
          <path 
            className="opacity-75" 
            fill="currentColor" 
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
      ) : (
        <>
          {children}
          <span className="absolute inset-0 rounded-lg bg-white/0 group-hover:bg-white/10 transition-colors duration-200" />
        </>
      )}
    </button>
  );
} 