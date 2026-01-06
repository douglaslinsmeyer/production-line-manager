import { SpeakerWaveIcon } from '@heroicons/react/24/outline';
import type { ProfileStateOutputs } from '../../api/types';

interface StateOutputPreviewProps {
  outputs: ProfileStateOutputs;
  className?: string;
}

export default function StateOutputPreview({ outputs, className = '' }: StateOutputPreviewProps) {
  const getLightClass = (mode: string, color: string) => {
    const baseColors = {
      red: 'bg-red-500',
      yellow: 'bg-yellow-400',
      green: 'bg-green-500',
    };

    const offColors = {
      red: 'bg-gray-300',
      yellow: 'bg-gray-300',
      green: 'bg-gray-300',
    };

    if (mode === 'off') {
      return `${offColors[color as keyof typeof offColors]} opacity-30`;
    }

    if (mode === 'on') {
      return baseColors[color as keyof typeof baseColors];
    }

    if (mode === 'shortBlink') {
      return `${baseColors[color as keyof typeof baseColors]} animate-pulse`;
    }

    if (mode === 'longBlink') {
      return `${baseColors[color as keyof typeof baseColors]} animate-[pulse_3s_ease-in-out_infinite]`;
    }

    return offColors[color as keyof typeof offColors];
  };

  const getBuzzerClass = () => {
    if (outputs.buzzer === 'off') {
      return 'text-gray-300';
    }
    if (outputs.buzzer === 'on') {
      return 'text-blue-600';
    }
    if (outputs.buzzer === 'chirp') {
      return 'text-blue-600 animate-pulse';
    }
    return 'text-gray-300';
  };

  return (
    <div className={`flex items-center gap-4 ${className}`}>
      {/* Tower Light Stack */}
      <div className="flex flex-col gap-1">
        <div
          className={`w-6 h-6 rounded-full border-2 border-gray-400 ${getLightClass(outputs.redLight, 'red')}`}
          title={`Red: ${outputs.redLight}`}
        />
        <div
          className={`w-6 h-6 rounded-full border-2 border-gray-400 ${getLightClass(outputs.yellowLight, 'yellow')}`}
          title={`Yellow: ${outputs.yellowLight}`}
        />
        <div
          className={`w-6 h-6 rounded-full border-2 border-gray-400 ${getLightClass(outputs.greenLight, 'green')}`}
          title={`Green: ${outputs.greenLight}`}
        />
      </div>

      {/* Buzzer Indicator */}
      <div className="flex flex-col items-center">
        <SpeakerWaveIcon
          className={`w-6 h-6 ${getBuzzerClass()}`}
          title={`Buzzer: ${outputs.buzzer}`}
        />
        <span className="text-xs text-gray-500 mt-1">{outputs.buzzer}</span>
      </div>
    </div>
  );
}
