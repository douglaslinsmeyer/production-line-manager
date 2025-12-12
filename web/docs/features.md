# Web UI Features

## Overview

The Assembly Line Manager web interface is a React-based dashboard for real-time monitoring and management of production lines, devices, schedules, and compliance metrics.

## Dashboard

**Real-time Production Line Monitoring**

- Live status indicators (on/off/maintenance/error)
- Auto-refresh every 5 seconds using TanStack Query
- Quick status change actions
- Summary statistics (total lines, active, offline, error)
- Color-coded status badges
- Responsive grid layout

**Features**:
- Filter by status or label
- Search by line code or name
- Sort by various fields
- Card or table view toggle

## Production Lines Management

**CRUD Operations**

- **Create**: Add new production lines with code, name, description
- **Read**: View line details with full status history
- **Update**: Edit line information
- **Delete**: Soft-delete lines (preserves history)

**Status Management**:
- Change status with single click
- Status options: on, off, maintenance, error
- Audit trail with timestamps and sources
- Timeline view of all changes

**Label System**:
- Assign multiple labels to lines
- Color-coded labels for organization
- Filter and group by labels
- Label management (create, edit, delete)

## Schedules Management

**Schedule Builder**

- **Weekly Templates**: Configure 7-day schedules (Sunday-Saturday)
- **Per-Day Settings**:
  - Working/non-working flag
  - Shift start and end times
  - Multiple breaks per day (lunch, coffee breaks)
- **Timezone Support**: Configure schedules for different locations

**Holidays**:
- Manual holiday entry (date + name)
- Auto-import from OpenHolidays API
- Bulk import for multiple holidays
- Suggested holidays by country and year

**Exceptions**:
- **Global Exceptions**: Apply to all lines using schedule
- **Line-Specific Exceptions**: Apply to individual lines
- Date range configuration
- Modified schedule for exception period

**Effective Schedule Viewer**:
- Calculate effective schedule for any date range
- Shows priority resolution (line exception > global exception > holiday > base)
- Visual calendar view

## Analytics Dashboard

**Aggregate Metrics**

- Total uptime hours across all lines
- Average uptime percentage
- Total downtime hours
- Mean Time To Repair (MTTR)
- Total interruptions count

**Per-Line Metrics**

- Sortable table with all lines
- Uptime hours and percentage per line
- Downtime analysis
- MTTR per line
- Current status

**Visualizations** (using Recharts):
- **Status Distribution**: Pie chart showing time in each status
- **Daily KPIs**: Line chart of daily performance
- **Timeline**: Bar chart of status changes over time
- **Trends**: Historical uptime trends

**Date Range Selection**:
- Presets: 24 hours, 7 days, 30 days
- Custom range picker
- Export to CSV/Excel

## Compliance Tracking

**Overall Compliance**

- Aggregate compliance percentage
- Total scheduled vs actual uptime hours
- Unplanned downtime tracking
- Maintenance hours summary
- Working days count

**Per-Line Compliance**

- Compliance table with all lines
- Color-coded compliance percentages
  - Green: >95% (excellent)
  - Yellow: 90-95% (good)
  - Orange: 80-90% (fair)
  - Red: <80% (poor)
- Sortable by compliance percentage
- Filter by compliance level

**Daily Compliance Breakdown**:
- Daily compliance chart
- Line chart showing compliance over time
- Identify patterns and trends
- Drill-down to specific dates

## Device Discovery & Management

**Device List**

- All discovered ESP32-S3 devices
- Real-time status (online/offline)
- Last seen timestamp
- Device capabilities (8DI/8DO, Ethernet/WiFi)
- Firmware version
- IP address and MAC address

**Device Actions**:
- **Flash Identify**: Trigger LED and buzzer for physical identification
- **Assign to Line**: Map device to production line
- **Unassign**: Remove device-to-line mapping
- **View Details**: Device info and capabilities

**Auto-Discovery**:
- Devices appear automatically when powered on
- Polling every 5 seconds for status updates
- Offline detection (2 minutes without heartbeat)

**Summary Statistics**:
- Total devices discovered
- Online devices count
- Assigned devices count
- Unassigned devices count

## User Interface

**Design System**:
- Tailwind CSS for styling
- Headless UI for accessible components
- Heroicons for consistent iconography
- Mobile-first responsive design

**Navigation**:
- Sidebar navigation
- Breadcrumbs for context
- Active route highlighting
- Collapsible on mobile

**Notifications**:
- Toast notifications for actions (react-hot-toast)
- Success/error feedback
- Auto-dismiss after 3 seconds
- Stack multiple notifications

**Form Handling**:
- React Hook Form for validation
- Zod schema validation
- Real-time error feedback
- Accessible form controls

## State Management

**TanStack Query** (formerly React Query):
- Server state management
- Automatic caching (5 minutes default)
- Auto-refetch strategies:
  - Dashboard: Every 5 seconds
  - Devices: Every 5 seconds
  - Analytics: On window focus
  - Schedules: Manual refetch
- Optimistic updates for mutations
- Error handling and retry logic

**Local State**:
- React useState for UI state
- Form state via React Hook Form
- No global state management needed

## API Integration

**Axios HTTP Client**:
- Base URL configuration via environment
- Automatic JSON serialization
- Error interceptors
- Request/response logging (dev mode)

**API Client Structure**:
```typescript
// api/lines.ts
export const getLines = () => axios.get('/lines')
export const updateLineStatus = (id, status) =>
  axios.post(`/lines/${id}/status`, { status })

// hooks/useLines.ts
export const useLines = () =>
  useQuery({ queryKey: ['lines'], queryFn: getLines })
```

## Performance Optimizations

**Code Splitting**:
- Route-based code splitting with React Router
- Lazy loading for heavy components
- Dynamic imports for analytics charts

**Memoization**:
- React.memo for expensive components
- useMemo for computed values
- useCallback for stable function references

**Virtual Scrolling**:
- Not yet implemented (consider for large line lists)

**Bundle Size**:
- Production build optimized with Vite
- Tree-shaking for unused code
- Gzip compression in production

## Development Experience

**Hot Module Replacement (HMR)**:
- Instant updates during development
- Preserves component state
- Fast refresh with Vite

**TypeScript**:
- Type safety across application
- IntelliSense in IDE
- Catch errors at compile time
- Shared types with API

**Dev Tools**:
- React DevTools for component debugging
- TanStack Query DevTools for cache inspection
- Vite dev server with instant updates

## Accessibility

**WCAG 2.1 AA Compliance**:
- Semantic HTML
- ARIA labels and roles
- Keyboard navigation
- Focus management
- Color contrast ratios

**Screen Reader Support**:
- Descriptive labels
- Status announcements
- Form error descriptions

## Tech Stack Summary

| Category | Technology | Purpose |
|----------|-----------|---------|
| Framework | React 18 | UI library |
| Language | TypeScript | Type safety |
| Build Tool | Vite | Dev server, bundling |
| Routing | React Router v6 | Client-side routing |
| State | TanStack Query | Server state |
| HTTP | Axios | API requests |
| Styling | Tailwind CSS | Utility-first CSS |
| Components | Headless UI | Accessible components |
| Charts | Recharts | Analytics visualization |
| Forms | React Hook Form | Form handling |
| Validation | Zod | Schema validation |
| Notifications | react-hot-toast | Toast messages |
| Icons | Heroicons | Icon library |

## Browser Support

- Chrome/Edge: Latest 2 versions
- Firefox: Latest 2 versions
- Safari: Latest 2 versions
- Mobile browsers: iOS Safari 12+, Chrome Android

## See Also

- [API Documentation](../../api/README.md)
- [Architecture Overview](../../docs/architecture.md)
- [Getting Started Guide](../../docs/getting-started.md)
