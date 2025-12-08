import { useState, useEffect } from 'react';

interface SidebarState {
  isCollapsed: boolean;
  isMobileOpen: boolean;
  toggleCollapsed: () => void;
  toggleMobileMenu: () => void;
  closeMobileMenu: () => void;
}

const SIDEBAR_STORAGE_KEY = 'sidebar-collapsed';
const LG_BREAKPOINT = 1024;

// Synchronously read from localStorage to prevent layout shift
const getInitialCollapsed = (): boolean => {
  if (typeof window === 'undefined') return false;
  try {
    const stored = localStorage.getItem(SIDEBAR_STORAGE_KEY);
    return stored === 'true';
  } catch {
    return false;
  }
};

export function useSidebar(): SidebarState {
  const [isCollapsed, setIsCollapsed] = useState(getInitialCollapsed);
  const [isMobileOpen, setIsMobileOpen] = useState(false);

  // Persist collapsed state to localStorage
  useEffect(() => {
    try {
      localStorage.setItem(SIDEBAR_STORAGE_KEY, String(isCollapsed));
    } catch (error) {
      console.error('Failed to save sidebar state:', error);
    }
  }, [isCollapsed]);

  // Close mobile menu on window resize to desktop
  useEffect(() => {
    const handleResize = () => {
      if (window.innerWidth >= LG_BREAKPOINT) {
        setIsMobileOpen(false);
      }
    };

    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  const toggleCollapsed = () => {
    setIsCollapsed((prev) => !prev);
  };

  const toggleMobileMenu = () => {
    setIsMobileOpen((prev) => !prev);
  };

  const closeMobileMenu = () => {
    setIsMobileOpen(false);
  };

  return {
    isCollapsed,
    isMobileOpen,
    toggleCollapsed,
    toggleMobileMenu,
    closeMobileMenu,
  };
}
