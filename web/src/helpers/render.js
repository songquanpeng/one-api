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
  let groups = group.split(",");
  groups.sort();
  return <>
    {groups.map((group) => {
      if (group === "vip" || group === "pro") {
        return <Label color='yellow'>{group}</Label>
      } else if (group === "svip" || group === "premium") {
        return <Label color='red'>{group}</Label>
      }
      return <Label>{group}</Label>
    })}
  </>
}