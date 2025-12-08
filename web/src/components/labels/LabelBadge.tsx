import { XMarkIcon } from '@heroicons/react/20/solid';
import type { Label } from '@/api/types';

interface LabelBadgeProps {
  label: Label;
  onRemove?: () => void;
  size?: 'sm' | 'md';
}

// Generate consistent color from label name using hash
function getColorForLabel(name: string): string {
  const colors = [
    'bg-blue-100 text-blue-700 border-blue-200',
    'bg-green-100 text-green-700 border-green-200',
    'bg-purple-100 text-purple-700 border-purple-200',
    'bg-pink-100 text-pink-700 border-pink-200',
    'bg-yellow-100 text-yellow-700 border-yellow-200',
    'bg-indigo-100 text-indigo-700 border-indigo-200',
    'bg-orange-100 text-orange-700 border-orange-200',
    'bg-teal-100 text-teal-700 border-teal-200',
  ];

  const hash = name.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0);
  return colors[hash % colors.length];
}

// Helper function to lighten a hex color for background
function lightenColor(hexColor: string, percent: number = 85): string {
  const hex = hexColor.replace('#', '');
  const r = parseInt(hex.substring(0, 2), 16);
  const g = parseInt(hex.substring(2, 4), 16);
  const b = parseInt(hex.substring(4, 6), 16);

  const lighten = (value: number) => Math.round(value + (255 - value) * (percent / 100));

  const newR = lighten(r).toString(16).padStart(2, '0');
  const newG = lighten(g).toString(16).padStart(2, '0');
  const newB = lighten(b).toString(16).padStart(2, '0');

  return `#${newR}${newG}${newB}`;
}

export default function LabelBadge({ label, onRemove, size = 'md' }: LabelBadgeProps) {
  const sizeClass = size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-2.5 py-1 text-sm';

  // Check if color is a hex value or Tailwind class
  const isHexColor = label.color?.startsWith('#');

  let colorClass = '';
  let inlineStyle: React.CSSProperties = {};

  if (isHexColor && label.color) {
    // Use inline styles for hex colors
    const bgColor = lightenColor(label.color);
    const textColor = label.color;
    const borderColor = label.color + '33'; // Add transparency to border

    inlineStyle = {
      backgroundColor: bgColor,
      color: textColor,
      borderColor: borderColor,
    };
  } else {
    // Use Tailwind classes
    colorClass = label.color || getColorForLabel(label.name);
  }

  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full border font-medium ${colorClass} ${sizeClass}`}
      style={isHexColor ? inlineStyle : undefined}
      title={label.description || label.name}
    >
      {label.name}
      {onRemove && (
        <button
          onClick={(e) => {
            e.stopPropagation();
            onRemove();
          }}
          className="hover:opacity-70 transition-opacity"
          aria-label={`Remove label ${label.name}`}
          type="button"
        >
          <XMarkIcon className="h-3 w-3" />
        </button>
      )}
    </span>
  );
}
