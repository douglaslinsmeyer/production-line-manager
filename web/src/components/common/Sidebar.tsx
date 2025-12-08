import { Fragment } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { Dialog, Transition } from '@headlessui/react';
import {
  HomeIcon,
  ChartBarIcon,
  TagIcon,
  CalendarDaysIcon,
  DocumentTextIcon,
  XMarkIcon,
  ChevronDoubleLeftIcon,
  ChevronDoubleRightIcon,
} from '@heroicons/react/24/outline';
import { getApiServerUrl } from '../../api/client';

interface NavItem {
  name: string;
  path: string;
  icon: React.ComponentType<{ className?: string }>;
  external?: boolean;
}

interface SidebarProps {
  isCollapsed: boolean;
  isMobileOpen: boolean;
  onToggleCollapsed: () => void;
  onCloseMobile: () => void;
}

const navigationItems: NavItem[] = [
  {
    name: 'Dashboard',
    path: '/',
    icon: HomeIcon,
  },
  {
    name: 'Analytics',
    path: '/analytics',
    icon: ChartBarIcon,
  },
  {
    name: 'Schedules',
    path: '/schedules',
    icon: CalendarDaysIcon,
  },
  {
    name: 'Labels',
    path: '/labels',
    icon: TagIcon,
  },
  {
    name: 'API Docs',
    path: `${getApiServerUrl()}/swagger/`,
    icon: DocumentTextIcon,
    external: true,
  },
];

export default function Sidebar({
  isCollapsed,
  isMobileOpen,
  onToggleCollapsed,
  onCloseMobile,
}: SidebarProps) {
  const location = useLocation();

  const isActive = (path: string) => {
    if (path === '/') {
      return location.pathname === '/';
    }
    return location.pathname.startsWith(path);
  };

  const NavItemContent = ({ item, collapsed = false }: { item: NavItem; collapsed?: boolean }) => {
    const active = !item.external && isActive(item.path);
    const Icon = item.icon;

    const className = `
      flex items-center rounded-lg transition-colors
      ${active ? 'bg-blue-50 text-blue-600' : 'text-gray-700 hover:bg-gray-100'}
      ${collapsed ? 'justify-center p-3' : 'gap-3 px-3 py-2'}
    `.trim();

    const content = (
      <>
        <Icon className="h-6 w-6 flex-shrink-0" />
        {!collapsed && <span className="font-medium">{item.name}</span>}
      </>
    );

    if (item.external) {
      return (
        <a
          href={item.path}
          target="_blank"
          rel="noopener noreferrer"
          className={className}
          aria-label={collapsed ? item.name : undefined}
          title={collapsed ? item.name : undefined}
        >
          {content}
        </a>
      );
    }

    return (
      <Link
        to={item.path}
        className={className}
        aria-current={active ? 'page' : undefined}
        aria-label={collapsed ? item.name : undefined}
        title={collapsed ? item.name : undefined}
      >
        {content}
      </Link>
    );
  };

  const SidebarContent = ({ collapsed = false, onItemClick }: { collapsed?: boolean; onItemClick?: () => void }) => (
    <div className="flex flex-col h-full">
      {/* Logo */}
      <Link
        to="/"
        className="flex items-center px-4 py-4 border-b border-gray-200"
        onClick={onItemClick}
      >
        {collapsed ? (
          <div className="h-8 w-8 bg-blue-600 rounded-lg flex items-center justify-center mx-auto">
            <span className="text-white font-bold text-lg">L</span>
          </div>
        ) : (
          <h1 className="text-lg font-bold text-gray-900">
            Line Manager
          </h1>
        )}
      </Link>

      {/* Navigation */}
      <nav aria-label="Main navigation" className="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
        {navigationItems.map((item) => (
          <div key={item.name} onClick={onItemClick}>
            <NavItemContent item={item} collapsed={collapsed} />
          </div>
        ))}
      </nav>

      {/* Collapse toggle (desktop only) */}
      {!isMobileOpen && (
        <div className="hidden lg:block border-t border-gray-200 p-3">
          <button
            onClick={onToggleCollapsed}
            aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
            aria-expanded={!collapsed}
            className={`
              flex items-center rounded-lg transition-colors text-gray-700 hover:bg-gray-100 w-full
              ${collapsed ? 'justify-center p-3' : 'gap-3 px-3 py-2'}
            `.trim()}
          >
            {collapsed ? (
              <ChevronDoubleRightIcon className="h-6 w-6" />
            ) : (
              <>
                <ChevronDoubleLeftIcon className="h-6 w-6" />
                <span className="font-medium">Collapse</span>
              </>
            )}
          </button>
        </div>
      )}
    </div>
  );

  return (
    <>
      {/* Desktop Sidebar */}
      <aside
        className={`
          hidden lg:flex flex-col fixed inset-y-0 left-0 bg-white border-r border-gray-200 transition-all duration-300 z-30
          ${isCollapsed ? 'w-16' : 'w-64'}
        `.trim()}
      >
        <SidebarContent collapsed={isCollapsed} />
      </aside>

      {/* Mobile Sidebar (Dialog Overlay) */}
      <Transition appear show={isMobileOpen} as={Fragment}>
        <Dialog as="div" className="relative z-50 lg:hidden" onClose={onCloseMobile}>
          {/* Backdrop */}
          <Transition.Child
            as={Fragment}
            enter="ease-out duration-300"
            enterFrom="opacity-0"
            enterTo="opacity-100"
            leave="ease-in duration-200"
            leaveFrom="opacity-100"
            leaveTo="opacity-0"
          >
            <div className="fixed inset-0 bg-black/30 backdrop-blur-sm" />
          </Transition.Child>

          {/* Sidebar panel */}
          <div className="fixed inset-0">
            <Transition.Child
              as={Fragment}
              enter="ease-out duration-300"
              enterFrom="-translate-x-full"
              enterTo="translate-x-0"
              leave="ease-in duration-200"
              leaveFrom="translate-x-0"
              leaveTo="-translate-x-full"
            >
              <Dialog.Panel className="fixed inset-y-0 left-0 w-64 bg-white border-r border-gray-200 shadow-xl transform transition-transform">
                {/* Close button */}
                <div className="absolute top-4 right-4">
                  <button
                    onClick={onCloseMobile}
                    className="text-gray-400 hover:text-gray-600 transition-colors"
                    aria-label="Close menu"
                  >
                    <XMarkIcon className="h-6 w-6" />
                  </button>
                </div>

                <SidebarContent onItemClick={onCloseMobile} />
              </Dialog.Panel>
            </Transition.Child>
          </div>
        </Dialog>
      </Transition>
    </>
  );
}
