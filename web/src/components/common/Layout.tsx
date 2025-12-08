import { useEffect, type ReactNode } from 'react';
import { useLocation } from 'react-router-dom';
import { useSidebar } from '../../hooks/useSidebar';
import MobileMenuButton from './MobileMenuButton';
import Sidebar from './Sidebar';

interface LayoutProps {
  children: ReactNode;
}

export default function Layout({ children }: LayoutProps) {
  const location = useLocation();
  const { isCollapsed, isMobileOpen, toggleCollapsed, toggleMobileMenu, closeMobileMenu } = useSidebar();

  // Close mobile menu on route change
  useEffect(() => {
    closeMobileMenu();
  }, [location.pathname, closeMobileMenu]);

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Mobile menu button */}
      <MobileMenuButton onClick={toggleMobileMenu} isOpen={isMobileOpen} />

      {/* Sidebar */}
      <Sidebar
        isCollapsed={isCollapsed}
        isMobileOpen={isMobileOpen}
        onToggleCollapsed={toggleCollapsed}
        onCloseMobile={closeMobileMenu}
      />

      {/* Main content */}
      <main
        className={`
          transition-all duration-300 min-h-screen
          ${isCollapsed ? 'lg:ml-16' : 'lg:ml-64'}
          pt-16 lg:pt-0
        `.trim()}
      >
        <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
          {children}
        </div>
      </main>
    </div>
  );
}
