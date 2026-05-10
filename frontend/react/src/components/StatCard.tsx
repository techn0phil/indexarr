import comStyles from '../styles/components.module.css';

interface StatCardProps {
  label: string;
  value: string | number;
  subLabels?: string[];
  error?: boolean;
}

export const StatCard = ({ label, value, subLabels = [], error }: StatCardProps) => {
  return (
    <div className={comStyles.stat}>
      <div className={comStyles['stat-label']}>{label}</div>
      <div className={comStyles['stat-value']} style={{ color: error ? '#E24B4A' : 'var(--color-text-primary)' }}>
        {value}
      </div>
      {subLabels.map((subLabel, index) => (
        <div key={index} className={comStyles['stat-sub']}>
          {subLabel}
        </div>
      ))}
    </div>
  );
};
