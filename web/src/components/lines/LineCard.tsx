import { Link } from 'react-router-dom';
import type { ProductionLine } from '@/api/types';
import Card from '@/components/common/Card';
import StatusCircle from '@/components/common/StatusCircle';
import StatusTimer from './StatusTimer';
import SegmentedStatusButton from './SegmentedStatusButton';

interface LineCardProps {
  line: ProductionLine;
}

export default function LineCard({ line }: LineCardProps) {
  return (
    <>
      <Card className="hover:shadow-md transition-shadow">
        <div className="flex items-center gap-4">
          {/* Left: Status Circle */}
          <div className="flex-shrink-0">
            <StatusCircle status={line.status} size="lg" />
          </div>

          {/* Center: Main Content */}
          <div className="flex-1 min-w-0">
            {/* Header Row with Name, Timer, and Segmented Button */}
            <div className="flex items-start justify-between gap-4">
              <div className="flex items-center gap-3 flex-1 min-w-0">
                <Link to={`/lines/${line.id}`}>
                  <h3 className="text-lg font-semibold text-gray-900 hover:text-blue-600 truncate">
                    {line.name}
                  </h3>
                </Link>
                <StatusTimer lineId={line.id} currentStatus={line.status} />
              </div>
              {/* Right: Segmented Status Button */}
              <div className="flex-shrink-0">
                <SegmentedStatusButton lineId={line.id} currentStatus={line.status} />
              </div>
            </div>

            {/* Code and Description Row */}
            <div className="flex items-center gap-4 text-sm text-gray-500 -mt-2">
              <span>Code: {line.code}</span>
              {line.description && (
                <>
                  <span className="text-gray-300">|</span>
                  <span className="flex-1 truncate">{line.description}</span>
                </>
              )}
            </div>
          </div>
        </div>
      </Card>
    </>
  );
}
