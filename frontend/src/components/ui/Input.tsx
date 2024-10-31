interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
}

export default function Input({ label, error, ...props }: InputProps) {
  return (
    <div className="space-y-1">
      <label className="block text-sm font-medium text-gray-200">
        {label}
      </label>
      <div className="relative">
        <input
          {...props}
          className={`
            w-full px-4 py-2.5 bg-gray-800/50 border rounded-lg
            focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500
            transition-all duration-200 ease-in-out
            outline-none text-white placeholder-gray-400
            ${error ? 'border-red-500' : 'border-gray-700/50'}
            ${props.disabled ? 'opacity-50 cursor-not-allowed' : ''}
            backdrop-blur-sm
          `}
        />
        {error && (
          <p className="absolute -bottom-5 left-0 text-xs text-red-400">
            {error}
          </p>
        )}
      </div>
    </div>
  );
} 