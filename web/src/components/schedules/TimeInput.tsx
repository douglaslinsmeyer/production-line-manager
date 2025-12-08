interface TimeInputProps {
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
  className?: string;
}

export default function TimeInput({
  value,
  onChange,
  disabled = false,
  className = '',
}: TimeInputProps) {
  // Convert HH:MM:SS to HH:MM for the input
  const displayValue = value ? value.substring(0, 5) : '';

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    // Convert HH:MM to HH:MM:SS for storage
    onChange(newValue ? `${newValue}:00` : '');
  };

  return (
    <input
      type="time"
      value={displayValue}
      onChange={handleChange}
      disabled={disabled}
      className={`w-full px-4 py-2 rounded-lg border border-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm disabled:bg-gray-100 disabled:text-gray-500 ${className}`}
    />
  );
}
