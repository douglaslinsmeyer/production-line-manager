import { Bars3Icon } from '@heroicons/react/24/outline';

interface MobileMenuButtonProps {
  onClick: () => void;
  isOpen: boolean;
}

export default function MobileMenuButton({ onClick, isOpen }: MobileMenuButtonProps) {
  return (
    <button
      onClick={onClick}
      aria-label="Open menu"
      aria-expanded={isOpen}
      className="fixed top-4 left-4 z-50 lg:hidden p-2 rounded-lg bg-white shadow-lg border border-gray-200 text-gray-700 hover:bg-gray-50 transition-colors"
    >
      <Bars3Icon className="h-6 w-6" />
    </button>
  );
}
