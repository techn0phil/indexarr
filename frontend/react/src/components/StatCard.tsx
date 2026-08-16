import comStyles from '../styles/components.module.css';

interface StatCardProps {
  label: string;
  value: string | number;
  subLabels?: string[];
  icon?: React.ReactNode;
  error?: boolean;
}

export const StatCard = ({ label, value, subLabels = [], icon, error }: StatCardProps) => {
  return (
    <div className={comStyles.stat}>
      {icon && <div className={comStyles['stat-watermark']}>{icon}</div>}
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
