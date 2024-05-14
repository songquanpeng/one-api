export const priceType = [
  { value: 'tokens', label: '按Token收费' },
  { value: 'times', label: '按次收费' }
];

export function ValueFormatter(value) {
  if (value == null) {
    return '';
  }
  if (value === 0) {
    return '免费';
  }
  return `$${parseFloat(value * 0.002).toFixed(6)} / ￥${parseFloat(value * 0.014).toFixed(6)}`;
}
