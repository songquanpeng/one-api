import { Label } from 'semantic-ui-react';

export function renderText(text, limit) {
  if (text.length > limit) {
    return text.slice(0, limit - 3) + '...';
  }
  return text;
}

export function renderGroup(group) {
  if (group === "") {
    return <Label>default</Label>
  }
  return <Label>{group}</Label>
}