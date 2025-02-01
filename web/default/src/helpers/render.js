import { Label } from 'semantic-ui-react';
import { useTranslation } from 'react-i18next';

export function renderText(text, limit) {
  if (text.length > limit) {
    return text.slice(0, limit - 3) + '...';
  }
  return text;
}

export function renderGroup(group) {
  if (group === '') {
    return <Label>default</Label>;
  }
  let groups = group.split(',');
  groups.sort();
  return (
    <>
      {groups.map((group) => {
        if (group === 'vip' || group === 'pro') {
          return <Label color='yellow'>{group}</Label>;
        } else if (group === 'svip' || group === 'premium') {
          return <Label color='red'>{group}</Label>;
        }
        return <Label>{group}</Label>;
      })}
    </>
  );
}

export function renderNumber(num) {
  if (num >= 1000000000) {
    return (num / 1000000000).toFixed(1) + 'B';
  } else if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M';
  } else if (num >= 10000) {
    return (num / 1000).toFixed(1) + 'k';
  } else {
    return num;
  }
}

export function renderQuota(quota, t, precision = 2) {
  const displayInCurrency = localStorage.getItem('display_in_currency') === 'true';
  const quotaPerUnit = parseFloat(localStorage.getItem('quota_per_unit') || '1');
  
  if (displayInCurrency) {
    const amount = (quota / quotaPerUnit).toFixed(precision);
    return t('common.quota.display_short', { amount });
  }
  
  return renderNumber(quota);
}

export function renderQuotaWithPrompt(quota, t) {
  const displayInCurrency = localStorage.getItem('display_in_currency') === 'true';
  const quotaPerUnit = parseFloat(localStorage.getItem('quota_per_unit') || '1');
  
  if (displayInCurrency) {
    const amount = (quota / quotaPerUnit).toFixed(2);
    return ` (${t('common.quota.display', { amount })})`;
  }
  
  return '';
}

const colors = [
  'red',
  'orange',
  'yellow',
  'olive',
  'green',
  'teal',
  'blue',
  'violet',
  'purple',
  'pink',
  'brown',
  'grey',
  'black',
];

export function renderColorLabel(text) {
  let hash = 0;
  for (let i = 0; i < text.length; i++) {
    hash = text.charCodeAt(i) + ((hash << 5) - hash);
  }
  let index = Math.abs(hash % colors.length);
  return (
    <Label basic color={colors[index]}>
      {text}
    </Label>
  );
}
